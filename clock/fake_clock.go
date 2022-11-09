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
	"sort"
	"sync"
	"time"

	"go.uber.org/atomic"
)

var _ Clock = (*FakeClock)(nil)

// A FakeClock is a manually-adjusted clock useful for mocking the flow of time
// in tests. It does not keep time by itself: use Add, SetTime, or SetNanotime
// functions to manage the clock's current time.
//
// Note that Timer- and Ticker-producing functions allocate internal types that
// are never freed.
type FakeClock struct {
	timers []*fakeTimer
	now    atomic.Int64
	mu     sync.Mutex
	clk    monotonicClock
}

// NewFakeClock creates a new FakeClock.
func NewFakeClock() *FakeClock {
	c := &FakeClock{}
	c.clk = monotonicClock{
		fn: func() int64 {
			return c.now.Load()
		},
	}
	return c
}

// Add adds the given time.Duration to the clock's internal time.
func (c *FakeClock) Add(d time.Duration) {
	c.checkTimers(c.now.Add(int64(d)))
}

// After returns a channel that receives the current time after d has elapsed.
func (c *FakeClock) After(d time.Duration) <-chan time.Time {
	return c.addTimer(d, nil).ch
}

// AfterFunc returns a timer that will invoke the given function after d has
// elapsed. The timer may be stopped and reset.
func (c *FakeClock) AfterFunc(d time.Duration, fn func()) *Timer {
	x := c.addTimer(d, fn)
	return &Timer{
		C:    x.ch,
		fake: x,
	}
}

// Nanotime returns the clock's internal time as integer nanoseconds.
func (c *FakeClock) Nanotime() int64 {
	return c.clk.Nanotime()
}

// NewTicker returns a new Ticker that receives time ticks every d. If d is not
// greater than zero, NewTicker will panic.
func (c *FakeClock) NewTicker(d time.Duration) *Ticker {
	if d <= 0 {
		panic("non-positive interval for FakeClock.NewTicker")
	}

	x := c.addTicker(d)
	return &Ticker{
		C:    x.ch,
		fake: x,
	}
}

// NewTimer returns a new Timer that receives a time tick after d.
func (c *FakeClock) NewTimer(d time.Duration) *Timer {
	x := c.addTimer(d, nil)
	return &Timer{
		C:    x.ch,
		fake: x,
	}
}

// Now returns the clock's internal time as time.Time.
func (c *FakeClock) Now() time.Time {
	return c.clk.Now()
}

// SetTime sets the clock to the given time.
func (c *FakeClock) SetTime(t time.Time) {
	c.SetNanotime(t.UnixNano())
}

// SetNanotime sets the clock to the given time in nanoseconds.
func (c *FakeClock) SetNanotime(n int64) {
	c.now.Store(n)
	c.checkTimers(n)
}

// Since returns the amount of time that elapsed between the clock's internal
// time and t.
func (c *FakeClock) Since(t time.Time) time.Duration {
	return c.SinceNanotime(t.UnixNano())
}

// SinceNanotime returns the amount of time that elapsed between the clock's
// internal time and ns.
func (c *FakeClock) SinceNanotime(ns int64) time.Duration {
	return time.Duration(c.clk.Nanotime() - ns)
}

// Sleep blocks for d.
//
// Note that Sleep must be called from a different goroutine than the clock's
// time is being managed on, or the program will deadlock.
func (c *FakeClock) Sleep(d time.Duration) {
	timer := c.addTimer(d, nil)
	defer c.removeTimer(timer)
	<-timer.ch
}

// NewStopwatch returns a new Stopwatch that uses the current clock for
// measuring time. The clock's current time is used as the stopwatch's epoch.
func (c *FakeClock) NewStopwatch() *Stopwatch {
	return newStopwatch(c)
}

// Tick returns a new channel that receives time ticks every d. It is
// equivalent to writing c.NewTicker(d).C(). The given duration must be greater
// than 0.
func (c *FakeClock) Tick(d time.Duration) <-chan time.Time {
	if d < 0 {
		panic("non-positive interval for FakeClock.Tick")
	}
	return c.NewTicker(d).C
}

func (c *FakeClock) addTicker(d time.Duration) *fakeTimer {
	fake := newFakeTicker(c, d)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.timers = append(c.timers, fake)
	c.sortTimersNosync()

	return fake
}

func (c *FakeClock) addTimer(d time.Duration, fn func()) *fakeTimer {
	fake := newFakeTimer(c, d, fn)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.timers = append(c.timers, fake)
	c.sortTimersNosync()

	return fake
}

func (c *FakeClock) checkTimers(now int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < len(c.timers); /* noincr */ {
		if when := c.timers[i].when; when < 0 || when > now {
			return
		}

		// This timer should tick. If it has a function, the function should be
		// called in its own goroutine; otherwise, the channel should receive a
		// tick.
		if c.timers[i].fn != nil {
			go c.timers[i].fn()
		} else {
			tick(c.timers[i].ch, c.timers[i].when)
		}

		// If this is a ticker, extend when by period.
		if c.timers[i].period != 0 {
			c.timers[i].when = now + c.timers[i].period
			i++
			continue
		}

		// Otherwise, remove the timer since it just fired.
		if i < len(c.timers)-1 {
			copy(c.timers[i:], c.timers[i+1:])
		}
		c.timers = c.timers[:len(c.timers)-1]
	}
}

func (c *FakeClock) resetTimer(fake *fakeTimer, d time.Duration) bool {
	now := fake.clk.Nanotime()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the timer exists using its previous value.
	pos := c.insertPosNosync(fake.when)

	fake.when = now + int64(d)
	if fake.period != 0 {
		fake.period = int64(d)
	}

	// The timer doesn't exist; insert it into its new position based on the
	// current time and given duration.
	if n := len(c.timers); n == 0 || pos >= n || c.timers[pos] != fake {
		c.timers = append(c.timers, fake)
		c.sortTimersNosync()
		return false
	}

	return true
}

func (c *FakeClock) removeTimer(fake *fakeTimer) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.timers) == 0 {
		return false
	}

	// Ensure that the timer exists using its previous value. If not, add it.
	pos := c.insertPosNosync(fake.when)
	if c.timers[pos] != fake {
		return false
	}

	if pos < len(c.timers)-1 {
		copy(c.timers[pos:], c.timers[pos+1:])
	}
	c.timers = c.timers[:len(c.timers)-1]

	return true
}

func (c *FakeClock) insertPosNosync(when int64) int {
	// Inline the stdlib search for parity. Ref:
	// https://cs.opensource.google/go/go/+/refs/tags/go1.18.1:src/sort/search.go;l=59-74
	i, j := 0, len(c.timers)
	for i < j {
		h := int(uint(i+j) >> 1)
		if cur := c.timers[i].when; cur >= 0 && cur < when {
			i = h + 1
		} else {
			j = h
		}
	}

	return i
}

func (c *FakeClock) sortTimersNosync() {
	sort.Slice(c.timers, func(i int, j int) bool {
		a, b := c.timers[i], c.timers[j]
		return b.when < 0 || (a.when >= 0 && a.when < b.when)
	})
}

type fakeTimer struct {
	clk    *FakeClock
	ch     chan time.Time
	fn     func() // timer only
	when   int64  // timer expiration or next tick
	period int64  // ticker only
}

func newFakeTimer(clk *FakeClock, d time.Duration, fn func()) *fakeTimer {
	return &fakeTimer{
		clk:  clk,
		ch:   make(chan time.Time, 1),
		fn:   fn,
		when: clk.Nanotime() + int64(d),
	}
}

func newFakeTicker(clk *FakeClock, d time.Duration) *fakeTimer {
	return &fakeTimer{
		clk:    clk,
		ch:     make(chan time.Time, 1),
		when:   clk.Nanotime() + int64(d),
		period: int64(d),
	}
}

func (f *fakeTimer) resetTimer(d time.Duration) bool {
	return f.clk.resetTimer(f, d)
}

func (f *fakeTimer) removeTimer() bool {
	return f.clk.removeTimer(f)
}

func tick(ch chan time.Time, ns int64) {
	ts := time.Unix(0, ns)
	for {
		select {
		case ch <- ts:
			return
		default:
			<-ch
		}
	}
}
