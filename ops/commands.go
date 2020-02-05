package ops

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/luno/fate"
	"github.com/luno/reflex"
	"github.com/luno/reflex/rpatterns"

	"exchange"
	"exchange/db/commands"
	"exchange/matcher"
)

type tempCursor struct {
	cursor string
}

func (m tempCursor) GetCursor(_ context.Context, _ string) (string, error) {
	return m.cursor, nil
}

func (m *tempCursor) SetCursor(_ context.Context, _ string, cursor string) error {
	(*m).cursor = cursor
	return nil
}

func (m tempCursor) Flush(_ context.Context) error {return nil}

type commandFeeder struct {
	dbc *sql.DB
	cmdFeed chan matcher.Command
}

func NewCommandFeed(dbc *sql.DB) *commandFeeder {
	return &commandFeeder{dbc: dbc}
}

func (f *commandFeeder) StreamCommands(ctx context.Context, from int64, cmdFeed chan matcher.Command) error {
	f.cmdFeed = cmdFeed
	// Reflex enqueues input from order events
	cs := &tempCursor{cursor: strconv.FormatInt(from, 10)}
	ac := rpatterns.NewAckConsumer("-", cs, f.Enqueue)
	spec := rpatterns.NewAckSpec(commands.ToStream(f.dbc), ac)

	return reflex.Run(ctx, spec)
}

func (f commandFeeder) Enqueue(_ context.Context, _ fate.Fate, e *rpatterns.AckEvent) error {
	var cmd matcher.Command

	switch {
	case reflex.IsType(e.Type, exchange.CommandPostLimitOrder):
		var o exchange.PostLimit
		err := json.Unmarshal(e.MetaData, &o)
		if err != nil {
			return err
		}
		t := matcher.CommandLimit
		if o.PostOnly {
			t = matcher.CommandPostOnly
		}
		cmd = matcher.Command{
			Type:          t,
			IsBuy:         o.IsBuy,
			OrderID:       e.ForeignIDInt(),
			LimitPrice:    o.LimitPrice,
			LimitVolume:   o.LimitVolume,
		}
	case reflex.IsType(e.Type, exchange.CommandPostMarketOrder):
		var o exchange.PostMarket
		err := json.Unmarshal(e.MetaData, &o)
		if err != nil {
			return err
		}
		cmd = matcher.Command{
			Type:          matcher.CommandMarket,
			IsBuy:         o.IsBuy,
			OrderID:       e.ForeignIDInt(),
			MarketBase:    o.MarketBase,
			MarketCounter: o.MarketCounter,
		}
	case reflex.IsType(e.Type, exchange.CommandStopOrder):
		var o exchange.StopOrder
		err := json.Unmarshal(e.MetaData, &o)
		if err != nil {
			return err
		}
		cmd = matcher.Command{
			Type:          matcher.CommandCancel,
			IsBuy:         o.IsBuy,
			OrderID:       e.ForeignIDInt(),
		}
	default:
		return nil
	}

	cmd.Sequence = e.IDInt()
	f.cmdFeed <- cmd
	return nil
}
