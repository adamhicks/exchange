package ops

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/luno/jettison/log"
	"github.com/luno/reflex"

	"exchange/db/cursors"
	"exchange/db/results"
	"exchange/matcher"
)

var resultCursor = "results_feed"

func CreateResultsLog(ctx context.Context, dbc *sql.DB, options ...Option) error {
	cs := cursors.ToStore(dbc)
	c, err := cs.GetCursor(ctx, resultCursor)
	if err != nil {
		return err
	}

	var seq int64
	if c != "" {
		seq, err = strconv.ParseInt(c, 10, 64)
		if err != nil {
			return err
		}
	}

	ms := NewMatcherStream(options...)
	out := make(chan matcher.Result, 100)
	go func() {
		defer close(out)
		err := ms.StreamMatcher(
			ctx,
			EmptySnapshot,
			NewCommandFeed(dbc).StreamCommands,
			seq,
			NewSnapshotter(dbc).StoreSnapshot,
			out,
		)
		if err != nil {
			log.Error(ctx, err)
		}
	}()
	return storeResultsFeed(ctx, dbc, out, seq, cs)
}

func storeResultsFeed(ctx context.Context, dbc *sql.DB, r chan matcher.Result, seq int64, cs reflex.CursorStore) error {
	for {
		// Read up to bax batch available results.
		var rl []matcher.Result
		for {
			// Pop a result if available on channel.
			var popped bool
			select {
			case r := <-r:
				rl = append(rl, r)
				popped = true
			default:
			}

			if !popped && len(rl) > 0 {
				// Nothing more available now, process batch.
				break
			} else if !popped && len(rl) == 0 {
				// Nothing available yet, wait a bit.
				time.Sleep(time.Millisecond)
				continue
			} else if popped && len(rl) >= 100 {
				// Max popped, process batch
				break
			} else /* popped && len(rl) < s.maxBatch */ {
				// Popped another, see if more available.
				continue
			}
		}

		var toStore []matcher.Result
		for _, r := range rl {
			if r.Sequence <= seq {
				continue
			}
			toStore = append(toStore, r)
		}

		if len(toStore) == 0 {
			continue
		}

		_, err := results.Create(ctx, dbc, toStore)
		if err != nil {
			return err
		}

		seq = toStore[len(toStore) - 1].Sequence
		err = cs.SetCursor(ctx, resultCursor, strconv.FormatInt(seq, 10))
		if err != nil {
			return err
		}
	}
}