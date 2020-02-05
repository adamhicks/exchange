package ops

import (
	"context"
	"database/sql"

	"github.com/luno/fate"
	"github.com/luno/reflex"
	"github.com/luno/reflex/rpatterns"

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
	if !reflex.IsAnyType(e.Type, orders.StatusPending, orders.StatusCancelling){
		return nil
	}

	tx, err := s.dbc.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if reflex.IsType(e.Type, orders.StatusPending) {
		n, err := commands.EnqueuePostOrder(ctx, tx, e.ForeignIDInt())
		if err != nil {
			return err
		}
		defer n()
	} else if reflex.IsType(e.Type, orders.StatusCancelling) {
		n, err := commands.EnqueueStopOrder(ctx, tx, e.ForeignIDInt())
		if err != nil {
			return err
		}
		defer n()
	}

	return tx.Commit()
}