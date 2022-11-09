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

// A Timer is functionally equivalent to a [time.Timer]. A Timer must be
// created by [Clock.NewTimer].
type Timer struct {
	C     <-chan time.Time
	timer *time.Timer
	fake  *fakeTimer
}

// Reset changes the timer to expire after duration d. It returns true if the
// timer had been active, false if the timer had expired or been stopped.
//
// See Reset documentation on [time.Timer] for more information.
func (t *Timer) Reset(d time.Duration) bool {
	if t.timer != nil {
		return t.timer.Reset(d)
	}

	return t.fake.resetTimer(d)
}

// Stop prevents the Timer from firing. It returns true if the call stops the
// timer, false if the timer has already expired or been stopped. Stop does not
// close the channel, to prevent a read from the channel succeeding
// incorrectly.
//
// See Stop documentation on [time.Timer] for more information.
func (t *Timer) Stop() bool {
	if t.timer != nil {
		return t.timer.Stop()
	}
	return t.fake.removeTimer()
}
