package trades

// Code generated by glean from glean.go:3. DO NOT EDIT.

import (
	"context"
	"database/sql"
)

const cols = " `id`, `seq`, `seq_idx`, `price`, `volume`, `maker_order_id`, `taker_order_id`, `created_at` "
const selectPrefix = "select " + cols + " from trades where "

func Lookup(ctx context.Context, dbc dbc, id int64) (*Trade, error) {
	return lookupWhere(ctx, dbc, "id=?", id)
}

// lookupWhere queries the trades table with the provided where clause, then scans
// and returns a single row.
func lookupWhere(ctx context.Context, dbc dbc, where string, args ...interface{}) (*Trade, error) {
	return scan(dbc.QueryRowContext(ctx, selectPrefix+where, args...))
}

// listWhere queries the trades table with the provided where clause, then scans
// and returns all the rows.
func listWhere(ctx context.Context, dbc dbc, where string, args ...interface{}) ([]Trade, error) {

	rows, err := dbc.QueryContext(ctx, selectPrefix+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Trade
	for rows.Next() {
		r, err := scan(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, *r)
	}

	return res, rows.Err()
}

func scan(row row) (*Trade, error) {
	var g glean

	err := row.Scan(&g.ID, &g.Seq, &g.SeqIdx, &g.Price, &g.Volume, &g.MakerOrderID, &g.TakerOrderID, &g.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &Trade{
		ID:           g.ID,
		Seq:          g.Seq,
		SeqIdx:       g.SeqIdx,
		Price:        g.Price,
		Volume:       g.Volume,
		MakerOrderID: g.MakerOrderID,
		TakerOrderID: g.TakerOrderID,
		CreatedAt:    g.CreatedAt,
	}, nil
}

// dbc is a common interface for *sql.DB and *sql.Tx.
type dbc interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// row is a common interface for *sql.Rows and *sql.Row.
type row interface {
	Scan(dest ...interface{}) error
}
