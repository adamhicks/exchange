package snapshots

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"

	"exchange/matcher"
)

func StoreOrderbook(ctx context.Context, db *sql.DB, book *matcher.OrderBook) error {
	d, err := json.Marshal(book)
	log.Info(ctx, "storing ss", j.KV("book", string(d)))
	if err != nil {
		return err
	}
	_, err = db.ExecContext(
		ctx,
		"insert into snapshots "+
		"(command_id, match_sequence, orderbook) "+
		"values "+
		"(?, ?, ?)",
		book.CommandSequence,
		book.ResultSequence,
		d,
	)
	return err
}