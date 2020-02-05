package ops

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/luno/fate"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/reflex"
	"github.com/luno/reflex/rpatterns"

	"exchange"
	"exchange/db/commands"
	"exchange/db/orders"
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

func (f commandFeeder) Enqueue(ctx context.Context, _ fate.Fate, e *rpatterns.AckEvent) error {
	var cmd matcher.Command

	switch {
	case reflex.IsType(e.Type, exchange.CommandTypePostOrder):
		c, err := makePostCommand(ctx, f.dbc, e.ForeignIDInt())
		if err != nil {
			return err
		}
		cmd = c
	case reflex.IsType(e.Type, exchange.CommandTypeStopOrder):
		c, err := makeStopCommand(ctx, f.dbc, e.ForeignIDInt())
		if err != nil {
			return err
		}
		cmd = c
	default:
		return nil
	}

	cmd.Sequence = e.IDInt()
	f.cmdFeed <- cmd
	return nil
}

func makePostCommand(ctx context.Context, dbc *sql.DB, orderID int64) (matcher.Command, error) {
	o, err := orders.Lookup(ctx, dbc, orderID)
	if err != nil {
		return matcher.Command{}, err
	}
	var m matcher.CommandType
	if o.Type == orders.TypeMarket {
		m = matcher.CommandMarket
	} else if o.Type == orders.TypePostOnly {
		m = matcher.CommandPostOnly
	} else if o.Type == orders.TypeLimit {
		m = matcher.CommandLimit
	} else {
		return matcher.Command{}, errors.New("invalid order type", j.KV("order_type", o.Type))
	}

	return matcher.Command{
		Type:          m,
		IsBuy:         o.IsBuy,
		OrderID:       o.ID,
		LimitPrice:    o.LimitPrice,
		LimitVolume:   o.LimitVolume,
		MarketBase:    o.MarketBase,
		MarketCounter: o.MarketCounter,
	}, nil
}

func makeStopCommand(ctx context.Context, dbc *sql.DB, orderID int64) (matcher.Command, error) {
	o, err := orders.Lookup(ctx, dbc, orderID)
	if err != nil {
		return matcher.Command{}, err
	}

	return matcher.Command{
		Type:          matcher.CommandCancel,
		IsBuy:         o.IsBuy,
		OrderID:       o.ID,
	}, nil
}