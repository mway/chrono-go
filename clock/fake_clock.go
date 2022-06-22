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
	hooks  resolvedHooks
}

// NewFakeClock creates a new FakeClock.
func NewFakeClock(opts ...FakeOption) *FakeClock {
	var (
		options = DefaultFakeOptions().With(opts...)
		c       = &FakeClock{
			hooks: newResolvedHooks(options.Hooks),
		}
	)

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
	return c.addTimer(int64(d), 0, nil).C()
}

// AfterFunc returns a timer that will invoke the given function after d has
// elapsed. The timer may be stopped and reset.
func (c *FakeClock) AfterFunc(d time.Duration, f func()) Timer {
	return c.addTimer(int64(d), 0, f)
}

// Nanotime returns the clock's internal time as integer nanoseconds.
func (c *FakeClock) Nanotime() int64 {
	return c.clk.Nanotime()
}

// NewTicker returns a new Ticker that receives time ticks every d. The given
// duration must be greater than 0.
func (c *FakeClock) NewTicker(d time.Duration) Ticker {
	if d <= 0 {
		panic("non-positive interval for FakeClock.NewTicker")
	}
	return fakeTicker{c.addTimer(int64(d), int64(d), nil)}
}

// NewTimer returns a new Timer that receives a time tick after d.
func (c *FakeClock) NewTimer(d time.Duration) Timer {
	return c.addTimer(int64(d), 0, nil)
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
	timer := c.addTimer(int64(d), 0, nil)
	defer func() {
		timer.Stop()
		c.removeTimer(timer)
	}()
	<-timer.ch
}

// Tick returns a new channel that receives time ticks every d. It is
// equivalent to writing c.NewTicker(d).C(). The given duration must be greater
// than 0.
func (c *FakeClock) Tick(d time.Duration) <-chan time.Time {
	if d < 0 {
		panic("non-positive interval for FakeClock.Tick")
	}
	return c.NewTicker(d).C()
}

func (c *FakeClock) addTimer(when int64, period int64, fn func()) *fakeTimer {
	dur := time.Duration(when)
	when = c.clk.Nanotime() + when

	c.mu.Lock()
	defer c.mu.Unlock()

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

	x := newFakeTimer(c, dur, when, period, fn)
	c.timers = append(c.timers, x)

	if i < len(c.timers) {
		copy(c.timers[i+1:], c.timers[i:])
		c.timers[i] = x
	}

	c.callCreateCallback(dur, period > 0)
	return c.timers[i]
}

func (c *FakeClock) callCreateCallback(dur time.Duration, ticker bool) {
	go func() {
		if ticker {
			c.hooks.OnTickerCreate(c, dur)
		} else {
			c.hooks.OnTimerCreate(c, dur)
		}
	}()
}

func (c *FakeClock) callResetCallback(dur time.Duration, ticker bool) {
	go func() {
		if ticker {
			c.hooks.OnTickerReset(c, dur)
		} else {
			c.hooks.OnTimerReset(c, dur)
		}
	}()
}

func (c *FakeClock) callStopCallback(ticker bool) {
	go func() {
		if ticker {
			c.hooks.OnTickerStop(c)
		} else {
			c.hooks.OnTimerStop(c)
		}
	}()
}

func (c *FakeClock) checkTimers(now int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var any bool
	defer func() {
		if any {
			c.sortTimersNosync()
		}
	}()

	for i := 0; i < len(c.timers); i++ {
		if when := c.timers[i].when; when < 0 || when > now {
			return
		}

		any = true
		c.timers[i].tick(now)
		if c.timers[i].period <= 0 {
			c.stopTimerNosync(c.timers[i])
		}
	}
}

func (c *FakeClock) removeTimer(t *fakeTimer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// We iterate backwards, because timers are currently only removed after
	// they have ticked.
	for i := len(c.timers) - 1; i >= 0; i-- {
		if t == c.timers[i] {
			copy(c.timers[i:], c.timers[i+1:])
			c.timers = c.timers[:len(c.timers)-1]
			return
		}
	}
}

func (c *FakeClock) resetTimer(t *fakeTimer, when int64) int64 {
	if t == nil {
		return 0
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.sortTimersNosync()

	prev := t.when
	t.when = c.clk.Nanotime() + when
	t.dur = time.Duration(when)

	if t.period > 0 {
		t.period = when
	}

	c.callResetCallback(t.dur, t.period > 0)
	return prev
}

func (c *FakeClock) sortTimersNosync() {
	sort.Slice(c.timers, func(i int, j int) bool {
		a, b := c.timers[i], c.timers[j]
		return b.when < 0 || (a.when >= 0 && a.when < b.when)
	})
}

func (c *FakeClock) stopTimer(t *fakeTimer) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stopTimerNosync(t)
}

func (c *FakeClock) stopTimerNosync(t *fakeTimer) int64 {
	if t == nil {
		return 0
	}

	for i := 0; i < len(c.timers); i++ {
		if t == c.timers[i] {
			defer c.callStopCallback(t.period > 0)

			if prev := t.when; prev >= 0 {
				t.when = -1
				return prev
			}

			break
		}
	}

	return 0
}
