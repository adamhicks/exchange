package main

import (
	"context"
	"database/sql"

	"exchange/db/cursors"
	"exchange/db/orders"
	"exchange/db/results"
	"exchange/db/trades"
	"exchange/matcher"
	"exchange/ops"

	"github.com/luno/fate"
	"github.com/luno/reflex"
)

// Run runs the exchange returning the first error.
func Run(ctx context.Context, dbc *sql.DB, opts ...ops.Option) error {
	var err error

	// Start the sequencer and result writer go routines, exit on first error.
	select {
	case err = <-goChan(func() error {
		// Reflex errors don't affect state, could actually just retry.
		return ops.CreateOrderCommands(ctx, dbc)
	}):
	case err = <-goChan(func() error {
		// State stores results. Errors restart matching since
		// result is lost.
		return ops.CreateResultsLog(ctx, dbc, opts...)
	}):
	}

	return err
}

func goChan(f func() error) <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- f()
		close(ch)
	}()
	return ch
}

func ConsumeResults(ctx context.Context, dbc *sql.DB) error {
	spec := reflex.NewSpec(
		results.ToStream(dbc),
		cursors.ToStore(dbc),
		makeResultConsumec(dbc),
	)
	return reflex.Run(ctx, spec)
}

func makeResultConsumec(dbc *sql.DB) reflex.Consumer {
	// These results always complete orders.
	complete := map[matcher.Type]bool{
		matcher.TypeLimitTaker:    true,
		matcher.TypeMarketEmpty:   true,
		matcher.TypeMarketPartial: true,
		matcher.TypeMarketFull:    true,
		matcher.TypeCancelled:     true,
	}

	posted := map[matcher.Type]bool{
		matcher.TypePosted:       true,
		matcher.TypeLimitMaker:   true,
		matcher.TypeLimitPartial: true,
	}

	return reflex.NewConsumer("result_consumer",
		func(ctx context.Context, f fate.Fate, e *reflex.Event) error {

			result, err := results.Lookup(ctx, dbc, e.ForeignIDInt())
			if err != nil {
				return err
			}

			for _, r := range result.Results {

				var completed []int64

				for i, t := range r.Trades {
					_, err := trades.Create(ctx, dbc, trades.CreateReq{
						IsBuy:        t.IsBuy,
						Seq:          r.Sequence,
						SeqIdx:       i,
						Price:        t.Price,
						Volume:       t.Volume,
						MakerOrderID: t.MakerOrderID,
						TakerOrderID: t.TakerOrderID,
					})
					// TODO(corver): Ignore duplicate on uniq index
					if err != nil {
						return err
					}

					if t.MakerFilled {
						completed = append(completed, t.MakerOrderID)
					}
				}

				if posted[r.Type] {
					err := orders.UpdatePosted(ctx, dbc, r.OrderID, r.Sequence)
					// TODO(corver): Ignore already posted errors
					if err != nil {
						return err
					}
				}

				if complete[r.Type] {
					completed = append(completed, r.OrderID)
				}

				for _, id := range completed {
					err := orders.Complete(ctx, dbc, id, r.Sequence)
					// TODO(corver): Ignore already complete errors
					if err != nil {
						return err
					}
				}
			}

			return nil
		},
	)
}
