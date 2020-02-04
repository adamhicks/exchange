package ops

import (
	"sync/atomic"
	"time"
)

type Metrics struct {
	count          int64 // Used with atomic
	latencyNanoSum int64
	latencyCount   int64
}

func (m *Metrics) Count() int64 {
	return atomic.LoadInt64(&m.count)
}
func (m *Metrics) incCount() {
	atomic.AddInt64(&m.count, 1)
}
func (m *Metrics) latency() func() {
	t0 := time.Now()
	return func() {
		nanos := time.Since(t0).Nanoseconds()
		atomic.AddInt64(&m.latencyNanoSum, nanos)
		atomic.AddInt64(&m.latencyCount, 1)
	}
}
func (m *Metrics) MeanLatency() time.Duration {
	nanos := atomic.LoadInt64(&m.latencyNanoSum)
	count := atomic.LoadInt64(&m.latencyCount)
	return time.Duration(nanos / count)
}
