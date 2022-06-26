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

// A Timer is functionally equivalent to a *time.Timer.
type Timer interface {
	// C returns a channel that receives a time tick once the Timer expires.
	C() <-chan time.Time
	// Reset resets the Timer to have the given expiration. It returns whether
	// the Timer was running when Reset was called.
	Reset(time.Duration) bool
	// Stop stops the Timer. It returns whether the Timer was running when Stop
	// was called.
	Stop() bool
}

type gotimer struct {
	*time.Timer
}

func (t gotimer) C() <-chan time.Time {
	return t.Timer.C
}

// A Ticker is functionally equivalent to a *time.Ticker.
type Ticker interface {
	// C returns a channel that receives time ticks on every interval.
	C() <-chan time.Time
	// Reset resets the Ticker to have the given interval. If d is not greater
	// than zero, Reset will panic.
	Reset(d time.Duration)
	// Stop stops the Ticker.
	Stop()
}

type goticker struct {
	*time.Ticker
}

func (t goticker) C() <-chan time.Time {
	return t.Ticker.C
}
