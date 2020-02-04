package commands

import (
	"context"
	"database/sql"
	"time"

	"github.com/luno/jettison/errors"
	"github.com/luno/reflex"
	"github.com/luno/reflex/rsql"

	"exchange"
)

var (
	events = rsql.NewEventsTableInt("command_events", rsql.WithEventsInMemNotifier())
)

func CreateCommand(ctx context.Context, dbc *sql.DB, c exchange.Command) error {
	tx, err := dbc.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	r, err := tx.ExecContext(
		ctx,
		"insert into commands " +
		"(type, order_id, created_at) " +
		"values (?, ?, ?)",
		c.Type, c.OrderId, time.Now(),
	)
	if err != nil {
		return err
	}
	count, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("failed to create new command row")
	}
	cmdID, err := r.LastInsertId()
	if err != nil {
		return err
	}
	notify, err := events.Insert(ctx, tx, cmdID, exchange.EventCommandCreated)
	if err != nil {
		return err
	}
	defer notify()
	return tx.Commit()
}

func ToStream(dbc *sql.DB) reflex.StreamFunc {
	return events.ToStream(dbc)
}