// Copyright (c) 2022 Matt Way
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE THE SOFTWARE.

package clock

import (
	"sync"
	"time"

	"go.mway.dev/chrono"
	"go.uber.org/atomic"
)

var _ Clock = (*ThrottledClock)(nil)

// A NanoFunc is a function that returns a nanosecond timestamp as an int64.
type NanoFunc = func() int64

// NewMonotonicNanoFunc returns a new, default NanoFunc that reports monotonic
// system time as nanoseconds.
func NewMonotonicNanoFunc() NanoFunc {
	return chrono.Nanotime
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
	wg       sync.WaitGroup
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

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.run(interval)
	}()

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

// After returns a channel that receives the current time after d has elapsed.
// This method is not throttled and uses Go's runtime timers.
func (c *ThrottledClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// AfterFunc returns a timer that will invoke the given function after d has
// elapsed. The timer may be stopped and reset. This method is not throttled
// and uses Go's runtime timers.
func (c *ThrottledClock) AfterFunc(d time.Duration, fn func()) *Timer {
	x := time.AfterFunc(d, fn)
	return &Timer{
		C:     x.C,
		timer: x,
	}
}

// Interval returns the interval at which the clock updates its internal time.
func (c *ThrottledClock) Interval() time.Duration {
	return c.interval
}

// Nanotime returns the current time as integer nanoseconds.
func (c *ThrottledClock) Nanotime() int64 {
	return c.now.Load()
}

// NewStopwatch returns a new Stopwatch that uses the current clock for
// measuring time. The clock's current time is used as the stopwatch's epoch.
func (c *ThrottledClock) NewStopwatch() *Stopwatch {
	return newStopwatch(c)
}

// NewTicker returns a new Ticker that receives time ticks every d. This method
// is not throttled and uses Go's runtime timers. If d is not greater than
// zero, NewTicker will panic.
func (c *ThrottledClock) NewTicker(d time.Duration) *Ticker {
	x := time.NewTicker(d)
	return &Ticker{
		C:      x.C,
		ticker: x,
	}
}

// NewTimer returns a new Timer that receives a time tick after d. This method
// is not throttled and uses Go's runtime timers.
func (c *ThrottledClock) NewTimer(d time.Duration) *Timer {
	x := time.NewTimer(d)
	return &Timer{
		C:     x.C,
		timer: x,
	}
}

// Now returns the current time as time.Time.
func (c *ThrottledClock) Now() time.Time {
	return time.Unix(0, c.now.Load())
}

// Since returns the amount of time that elapsed between the clock's internal
// time and t.
func (c *ThrottledClock) Since(t time.Time) time.Duration {
	return c.SinceNanotime(t.UnixNano())
}

// SinceNanotime returns the amount of time that elapsed between the clock's
// internal time and ns.
func (c *ThrottledClock) SinceNanotime(ns int64) time.Duration {
	return time.Duration(c.Nanotime() - ns)
}

// SinceTimestamp returns the time elapsed since ts. It is shorthand for
// Timestamp()-ts.
func (c *ThrottledClock) SinceTimestamp(ts chrono.Timestamp) time.Duration {
	return time.Duration(c.Timestamp() - ts)
}

// Sleep puts the current goroutine to sleep for d. This method is not
// throttled and uses Go's runtime timers.
func (c *ThrottledClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

// Stop stops the clock. Note that this has no effect on currently-running
// timers.
func (c *ThrottledClock) Stop() {
	if c.stopped.CAS(false, true) {
		close(c.done)
	}
	c.wg.Wait()
}

// Tick returns a new channel that receives time ticks every d. It is
// equivalent to writing c.NewTicker(d).C().
func (c *ThrottledClock) Tick(d time.Duration) <-chan time.Time {
	//nolint:staticcheck
	return time.Tick(d)
}

// Timestamp returns the current time in nanoseconds as a
// [chrono.Timestamp].
func (c *ThrottledClock) Timestamp() chrono.Timestamp {
	return chrono.Timestamp(c.now.Load())
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
