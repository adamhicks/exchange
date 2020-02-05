package commands

import (
	"context"
	"database/sql"

	"github.com/luno/reflex"
	"github.com/luno/reflex/rsql"

	"exchange"
)

var (
	events = rsql.NewEventsTableInt("command_events", rsql.WithEventsInMemNotifier())
)

func EnqueuePostOrder(ctx context.Context, tx *sql.Tx, orderID int64) (rsql.NotifyFunc, error) {
	return events.Insert(ctx, tx, orderID, exchange.CommandTypePostOrder)
}

func EnqueueStopOrder(ctx context.Context, tx *sql.Tx, orderID int64) (rsql.NotifyFunc, error) {
	return events.Insert(ctx, tx, orderID, exchange.CommandTypeStopOrder)
}

func ToStream(dbc *sql.DB) reflex.StreamFunc {
	return events.ToStream(dbc)
}

func FillGaps(dbc *sql.DB) {
	rsql.FillGaps(dbc, events)
}
