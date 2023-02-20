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

package clock

import (
	"time"

	"go.mway.dev/chrono"
)

var _ Clock = (*wallClock)(nil)

type wallClock struct {
	fn TimeFunc
}

func newWallClock(fn TimeFunc) *wallClock {
	return &wallClock{
		fn: fn,
	}
}

func (c *wallClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (c *wallClock) AfterFunc(d time.Duration, fn func()) *Timer {
	x := time.AfterFunc(d, fn)
	return &Timer{
		C:     x.C,
		timer: x,
	}
}

func (c *wallClock) Nanotime() int64 {
	return c.fn().UnixNano()
}

func (c *wallClock) NewStopwatch() *Stopwatch {
	return newStopwatch(c)
}

func (c *wallClock) NewTicker(d time.Duration) *Ticker {
	ticker := time.NewTicker(d)
	return &Ticker{
		C:      ticker.C,
		ticker: ticker,
	}
}

func (c *wallClock) NewTimer(d time.Duration) *Timer {
	x := time.NewTimer(d)
	return &Timer{
		C:     x.C,
		timer: x,
	}
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

func (c *wallClock) SinceTimestamp(ts chrono.Timestamp) time.Duration {
	return time.Duration(c.Timestamp() - ts)
}

func (c *wallClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (c *wallClock) Tick(d time.Duration) <-chan time.Time {
	//nolint:staticcheck
	return time.Tick(d)
}

func (c *wallClock) Timestamp() chrono.Timestamp {
	return chrono.Timestamp(c.fn().UnixNano())
}
