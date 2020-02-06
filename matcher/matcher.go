package matcher

import (
	"context"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
)

// Match applies commands to the order book and outputs
// results including trades. The order book and input commands
// should be sequential. The snap function allows taking
// snapshots of the order book. The latency function allows
// measuring match latency.
func Match(ctx context.Context, book OrderBook,
	input <-chan Command, output chan<- Result,
	scale int, snap func(*OrderBook), latency func() func()) error {

	for {
		var cmd Command
		select {
		case <-ctx.Done():
			return ctx.Err()
		case cmd = <-input:
		}

		if cmd.Sequence <= book.CommandSequence {
			// Ignore old commands
			log.Info(ctx, "ignoring old command", j.KV("command_seq", cmd.Sequence))
			continue
		} else if cmd.Sequence > book.CommandSequence + 1 {
			return errors.New("out of order command",
				j.MKV{"expect": book.CommandSequence + 1, "got": cmd.Sequence})
		}

		l := latency()
		typ, tl := match(&book, cmd, scale)
		l()

		book.CommandSequence = cmd.Sequence
		book.ResultSequence++

		output <- Result{
			Sequence: book.ResultSequence,
			OrderID:  cmd.OrderID,
			Type:     typ,
			Trades:   tl,
		}

		// Call some metrics
		snap(&book)
	}
}
