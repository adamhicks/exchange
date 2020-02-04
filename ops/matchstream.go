package ops

import (
	"context"
	"sync"

	"github.com/luno/jettison/log"

	"exchange/matcher"
)

type SnapshotProvider func (ctx context.Context, before int64) (matcher.OrderBook, error)
type CommandStream func (ctx context.Context, from int64, into chan matcher.Command) error

type Option func(*MatcherStream)

func WithMetrics(m *Metrics) Option {
	return func(s *MatcherStream) {
		s.incCount = m.incCount
		s.latency = m.latency
	}
}

type MatcherStream struct {
	snapMu sync.Mutex
	snapshot matcher.OrderBook

	incCount func()
	latency func() func()
}

func NewMatcherStream(options ...Option) *MatcherStream {
	s := &MatcherStream{
		incCount: func() {},
		latency: func() func() {return func(){}},
	}
	for _, o := range options {
		o(s)
	}
	return s
}

func (ms *MatcherStream) updateSnapshot(book *matcher.OrderBook){
	ms.snapMu.Lock()
	defer ms.snapMu.Unlock()

	ms.incCount()
	ms.snapshot = *book
}

func (ms *MatcherStream) StreamMatcher(
	ctx context.Context, ss SnapshotProvider, cs CommandStream,
	from int64, results chan matcher.Result,
) error {
	ob, err := ss(ctx, from)
	if err != nil {
		return err
	}
	in := make(chan matcher.Command, 1000)
	go func() {
		defer close(in)
		err = cs(ctx, ob.Sequence, in)
		if err != nil {
			log.Error(ctx, err)
		}
	}()
	return matcher.Match(ctx, ob, in, results, 8, ms.updateSnapshot, ms.latency)
}