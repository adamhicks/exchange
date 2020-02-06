package ops

import (
	"context"
	"database/sql"

	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"

	"exchange/db/snapshots"
	"exchange/matcher"
)

func EmptySnapshot(_ context.Context, _ int64) (matcher.OrderBook, error) {
	return matcher.OrderBook{}, nil
}

type Snapshotter struct {
	dbc *sql.DB
	lastShot int64
	everyN int64
}

func NewSnapshotter(dbc *sql.DB) *Snapshotter {
	return &Snapshotter{dbc: dbc, lastShot: 0, everyN: 10}
}

func (ss Snapshotter) StoreSnapshot(book *matcher.OrderBook) {
	ctx := log.ContextWith(context.Background(), j.KV("command_id", book.CommandSequence))
	if book.CommandSequence % ss.everyN != 0 {
		return
	}
	err := snapshots.StoreOrderbook(ctx, ss.dbc, book)
	if err != nil {
		log.Error(ctx, err)
	}
}