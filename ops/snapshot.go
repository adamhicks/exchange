package ops

import (
	"context"

	"exchange/matcher"
)

func EmptySnapshot(_ context.Context, _ int64) (matcher.OrderBook, error) {
	return matcher.OrderBook{}, nil
}
