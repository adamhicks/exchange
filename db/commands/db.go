package commands

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/reflex"
	"github.com/luno/reflex/rsql"

	"exchange"
	"exchange/db/orders"
)

var events = rsql.NewEventsTableInt(
	"command_events",
	rsql.WithEventsInMemNotifier(),
	rsql.WithEventMetadataField("metadata"),
)

func EnqueuePostOrder(ctx context.Context, tx *sql.Tx, orderID int64) (rsql.NotifyFunc, error) {
	o, err := orders.Lookup(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}

	if o.Type == orders.TypeLimit || o.Type == orders.TypePostOnly {
		return enqueueLimitOrder(ctx, tx, o)
	} else if o.Type == orders.TypeMarket {
		return enqueueMarketOrder(ctx, tx, o)
	} else {
		return nil, errors.New("unrecognised order type", j.MKV{
			"order_id": o.ID,
			"type": o.Type,
		})
	}
}

func enqueueLimitOrder(ctx context.Context, tx *sql.Tx, o *orders.Order) (rsql.NotifyFunc, error) {
	var po bool
	if o.Type == orders.TypePostOnly {
		po = true
	}
	c := exchange.PostLimit{
		IsBuy:         o.IsBuy,
		PostOnly:      po,
		LimitPrice:    o.LimitPrice,
		LimitVolume:   o.LimitVolume,
	}
	d, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return events.InsertWithMetadata(ctx, tx, o.ID, exchange.CommandPostLimitOrder, d)
}

func enqueueMarketOrder(ctx context.Context, tx *sql.Tx, o *orders.Order) (rsql.NotifyFunc, error) {
	c := exchange.PostMarket{
		IsBuy:         o.IsBuy,
		MarketBase:    o.MarketBase,
		MarketCounter: o.MarketCounter,
	}
	d, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return events.InsertWithMetadata(ctx, tx, o.ID, exchange.CommandPostMarketOrder, d)
}

func EnqueueStopOrder(ctx context.Context, tx *sql.Tx, orderID int64) (rsql.NotifyFunc, error) {
	o, err := orders.Lookup(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}
	c := exchange.StopOrder{IsBuy: o.IsBuy}
	d, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return events.InsertWithMetadata(ctx, tx, orderID, exchange.CommandStopOrder, d)
}

func ToStream(dbc *sql.DB) reflex.StreamFunc {
	return events.ToStream(dbc)
}

func FillGaps(dbc *sql.DB) {
	rsql.FillGaps(dbc, events)
}
