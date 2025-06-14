// Copyright (c) 2023 Matt Way
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

// Package clock provides clock-related types and utilities.
package clock

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrNoClockFunc is returned when creating a new [Clock] without a valid
	// [TimeFunc] or [NanotimeFunc].
	ErrNoClockFunc = errors.New("no clock function provided")

	_ Clock = (*monotonicClock)(nil)
	_ Clock = (*wallClock)(nil)
)

// A Clock tells time.
type Clock interface {
	// After waits for the duration to elapse and then sends the current time
	// on the returned channel. It is equivalent to NewTimer(d).C. The
	// underlying [Timer] is not recovered by the garbage collector until the
	// it fires. If efficiency is a concern, use [NewTimer] instead and call
	// [Timer.Stop] if the timer is no longer needed.
	After(d time.Duration) <-chan time.Time

	// AfterFunc waits for the duration to elapse and then calls fn in its own
	// goroutine. It returns a [Timer] that can be used to cancel the call using
	// its [Timer.Stop] method.
	AfterFunc(d time.Duration, fn func()) *Timer

	// Nanotime returns the current time in nanoseconds.
	Nanotime() int64

	// NewStopwatch returns a new [Stopwatch] that uses the [Clock] for
	// measuring time.
	NewStopwatch() *Stopwatch

	// NewTicker returns a new [Ticker] containing a channel that will send the
	// current time on the channel after each tick. The period of the ticks is
	// specified by the duration argument. The ticker will adjust the time
	// interval or drop ticks to make up for slow receivers. The duration d
	// must be greater than zero; if not, NewTicker will panic. Stop the ticker
	// to release associated resources.
	NewTicker(d time.Duration) *Ticker

	// NewTimer creates a new [Timer] that will send the current time on its
	// channel after at least d has elapsed.
	NewTimer(d time.Duration) *Timer

	// Now returns the current time. For wall clocks, this is the local time;
	// for monotonic clocks, this is the system's monotonic time. Other Clock
	// implementations may have different locale or clock time semantics.
	Now() time.Time

	// Since returns the time elapsed since t. It is shorthand for
	// Now().Sub(t).
	Since(t time.Time) time.Duration

	// SinceNanotime returns the time elapsed since ns. It is shorthand for
	// Nanotime()-ns.
	SinceNanotime(ns int64) time.Duration

	// Sleep pauses the current goroutine for at least the duration d. A
	// negative or zero duration causes Sleep to return immediately.
	Sleep(d time.Duration)

	// Tick is a convenience wrapper for [NewTicker] providing access to the
	// ticking channel only. While Tick is useful for clients that have no need
	// to shut down the [Ticker], be aware that without a way to shut it down
	// the underlying Ticker cannot be recovered by the garbage collector; it
	// "leaks". Like [NewTicker], Tick will panic if d <= 0.
	Tick(time.Duration) <-chan time.Time
}

// NewClock returns a new [Clock] based on the given options.
func NewClock(opts ...Option) (Clock, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt.apply(&options)
	}

	if options.NanotimeFunc != nil {
		return newMonotonicClock(options.NanotimeFunc), nil
	}

	return newWallClock(options.TimeFunc), nil
}

// MustClock panics if the given error is not nil, otherwise it returns the
// given [Clock].
func MustClock(clock Clock, err error) Clock {
	if err != nil {
		panic(fmt.Errorf("clock.MustClock received an error: %w", err))
	}
	return clock
}

// NewMonotonicClock returns a new monotonic [Clock].
func NewMonotonicClock() Clock {
	return MustClock(NewClock(WithNanotimeFunc(DefaultNanotimeFunc())))
}

// NewWallClock returns a new wall [Clock].
func NewWallClock() Clock {
	return MustClock(NewClock(WithTimeFunc(DefaultTimeFunc())))
}

type monotonicClock struct {
	baseClock

	fn NanotimeFunc
}

func newMonotonicClock(fn NanotimeFunc) *monotonicClock {
	return &monotonicClock{
		fn: fn,
	}
}

func (c *monotonicClock) Nanotime() int64 {
	return c.fn()
}

func (c *monotonicClock) NewStopwatch() *Stopwatch {
	return newStopwatch(c)
}

func (c *monotonicClock) Now() time.Time {
	return time.Unix(0, c.fn())
}

func (c *monotonicClock) Since(t time.Time) time.Duration {
	return c.SinceNanotime(t.UnixNano())
}

func (c *monotonicClock) SinceNanotime(ns int64) time.Duration {
	return time.Duration(c.Nanotime() - ns)
}

type wallClock struct {
	baseClock

	fn TimeFunc
}

func newWallClock(fn TimeFunc) *wallClock {
	return &wallClock{
		fn: fn,
	}
}

func (c *wallClock) Nanotime() int64 {
	return c.fn().UnixNano()
}

func (c *wallClock) NewStopwatch() *Stopwatch {
	return newStopwatch(c)
}

func (c *wallClock) Now() time.Time {
	return c.fn()
}

func (c *wallClock) Since(t time.Time) time.Duration {
	return c.SinceNanotime(t.UnixNano())
}

func (c *wallClock) SinceNanotime(ts int64) time.Duration {
	return time.Duration(c.Nanotime() - ts)
}

type baseClock struct{}

func (baseClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (baseClock) AfterFunc(d time.Duration, fn func()) *Timer {
	x := time.AfterFunc(d, fn)
	return &Timer{
		C:     x.C,
		timer: x,
	}
}

func (baseClock) NewTicker(d time.Duration) *Ticker {
	ticker := time.NewTicker(d)
	return &Ticker{
		C:      ticker.C,
		ticker: ticker,
	}
}

func (baseClock) NewTimer(d time.Duration) *Timer {
	x := time.NewTimer(d)
	return &Timer{
		C:     x.C,
		timer: x,
	}
}

func (baseClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (baseClock) Tick(d time.Duration) <-chan time.Time {
	return time.Tick(d)
}
