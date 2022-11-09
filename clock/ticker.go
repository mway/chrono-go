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
	"errors"
	"time"
)

// A Ticker is functionally equivalent to a [time.Ticker]. A Ticker must be
// created by [Clock.NewTicker].
type Ticker struct {
	C      <-chan time.Time
	ticker *time.Ticker
	fake   *fakeTimer
}

// Reset stops a ticker and resets its period to the specified duration. The
// next tick will arrive after the new period elapses. The duration d must be
// greater than zero; if not, Reset will panic.
func (t *Ticker) Reset(d time.Duration) {
	if t.ticker != nil {
		t.ticker.Reset(d)
		return
	}

	if d <= 0 {
		panic(errors.New("non-positive interval for Ticker.Reset"))
	}

	t.fake.resetTimer(d)
}

// Stop turns off a ticker. After Stop, no more ticks will be sent. Stop does
// not close the channel, to prevent a concurrent goroutine reading from the
// channel from seeing an erroneous "tick".
func (t *Ticker) Stop() {
	if t.ticker != nil {
		t.ticker.Stop()
		return
	}

	t.fake.removeTimer()
}
