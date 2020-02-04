package ops

import (
	"context"
	"database/sql"

	"github.com/luno/fate"
	"github.com/luno/reflex"
	"github.com/luno/reflex/rpatterns"

	"exchange"
	"exchange/db/commands"
	"exchange/db/cursors"
	"exchange/db/orders"
)

type sequencer struct {
	dbc *sql.DB
}

// CreateOrderCommands creates matching commands for order operations
func CreateOrderCommands(ctx context.Context, dbc *sql.DB) error {
	// Reflex enqueues input from order events
	cs := cursors.ToStore(dbc)
	name := "order_command_creation"
	sequence := sequencer{dbc: dbc}
	ac := rpatterns.NewAckConsumer(name, cs, sequence.EnqueueOrderCommand)
	spec := rpatterns.NewAckSpec(orders.ToStream(dbc), ac)

	return reflex.Run(ctx, spec)
}

func (s sequencer) EnqueueOrderCommand(ctx context.Context, _ fate.Fate, e *rpatterns.AckEvent) error {
	var t exchange.CommandType
	if reflex.IsType(e.Type, orders.StatusPending) {
		t = exchange.CommandTypePostOrder
	} else if reflex.IsType(e.Type, orders.StatusCancelling) {
		t = exchange.CommandTypeStopOrder
	} else {
		// We only care about pending and cancelling states.
		return nil
	}

	next := exchange.Command{
		Type:     t,
		OrderId:  e.ForeignIDInt(),
	}
	return commands.CreateCommand(ctx, s.dbc, next)
}