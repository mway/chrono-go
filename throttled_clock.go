package chrono

import (
	"time"

	"go.uber.org/atomic"
)

// A NanoFunc is a function that returns a nanosecond timestamp as an int64.
type NanoFunc = func() int64

// NewMonotonicNanoFunc returns a new, default NanoFunc that reports monotonic
// system time as nanoseconds.
func NewMonotonicNanoFunc() NanoFunc {
	return Nanotime
}

// NewWallNanoFunc returns a new, default NanoFunc that reports wall time as
// nanoseconds.
func NewWallNanoFunc() NanoFunc {
	return func() int64 {
		return time.Now().UnixNano()
	}
}

// ThrottledClock provides a simple interface to memoize repeated time syscalls
// within a given threshold.
type ThrottledClock struct {
	nowfn    NanoFunc
	done     chan struct{}
	now      atomic.Int64
	stopped  atomic.Bool
	interval time.Duration
}

// NewThrottledClock creates a new ThrottledClock that uses the given NanoFunc
// to update its internal time at the given interval. A ThrottledClock should
// be stopped via ThrottledClock.Stop once it is no longer used.
//
// Note that interval should be tuned to be greater than the actual frequency
// of calls to ThrottledClock.Nanos or ThrottledClock.Now (otherwise the clock
// will generate more time calls than it is saving).
func NewThrottledClock(nowfn NanoFunc, interval time.Duration) *ThrottledClock {
	c := &ThrottledClock{
		nowfn:    nowfn,
		done:     make(chan struct{}),
		interval: interval,
	}

	// Set the clock to an initial time value.
	c.now.Store(c.nowfn())

	go c.run(interval)
	return c
}

// NewThrottledMonotonicClock creates a new ThrottledClock that uses
// NewMonotonicNanoFunc as its backing time function. See NewThrottledClock for
// more information.
func NewThrottledMonotonicClock(interval time.Duration) *ThrottledClock {
	return NewThrottledClock(NewMonotonicNanoFunc(), interval)
}

// NewThrottledWallClock creates a new ThrottledClock that uses NewWallNanoFunc
// as its backing time function. See NewThrottledClock for more information.
func NewThrottledWallClock(interval time.Duration) *ThrottledClock {
	return NewThrottledClock(NewWallNanoFunc(), interval)
}

// Interval returns the interval at which the clock updates its internal time.
func (c *ThrottledClock) Interval() time.Duration {
	return c.interval
}

// Nanos returns the current time as integer nanoseconds.
func (c *ThrottledClock) Nanos() int64 {
	return c.now.Load()
}

// Now returns the current time as time.Time.
func (c *ThrottledClock) Now() time.Time {
	return time.Unix(0, c.now.Load())
}

// Stop stops the clock.
func (c *ThrottledClock) Stop() {
	if !c.stopped.CAS(false, true) {
		return
	}
	close(c.done)
}

func (c *ThrottledClock) run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.now.Store(c.nowfn())
		}
	}
}
