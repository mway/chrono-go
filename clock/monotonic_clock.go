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
	"time"
)

var _ Clock = (*monotonicClock)(nil)

type monotonicClock struct {
	fn NanotimeFunc
}

func newMonotonicClock(fn NanotimeFunc) *monotonicClock {
	return &monotonicClock{
		fn: fn,
	}
}

func (c *monotonicClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (c *monotonicClock) AfterFunc(d time.Duration, fn func()) *Timer {
	x := time.AfterFunc(d, fn)
	return &Timer{
		C:     x.C,
		timer: x,
	}
}

func (c *monotonicClock) Nanotime() int64 {
	return c.fn()
}

func (c *monotonicClock) NewStopwatch() *Stopwatch {
	return newStopwatch(c)
}

func (c *monotonicClock) NewTicker(d time.Duration) *Ticker {
	x := time.NewTicker(d)
	return &Ticker{
		C:      x.C,
		ticker: x,
	}
}

func (c *monotonicClock) NewTimer(d time.Duration) *Timer {
	timer := time.NewTimer(d)
	return &Timer{
		C:     timer.C,
		timer: timer,
	}
}

func (c *monotonicClock) Now() time.Time {
	return time.Unix(0, c.fn())
}

func (c *monotonicClock) Since(t time.Time) time.Duration {
	return c.SinceNanotime(t.UnixNano())
}

func (c *monotonicClock) SinceNanotime(t int64) time.Duration {
	return time.Duration(c.fn() - t)
}

func (c *monotonicClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (c *monotonicClock) Tick(d time.Duration) <-chan time.Time {
	//nolint:staticcheck
	return time.Tick(d)
}
