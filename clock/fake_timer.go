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

var _ Timer = (*fakeTimer)(nil)

type fakeTicker struct {
	*fakeTimer
}

func (t fakeTicker) Reset(d time.Duration) {
	t.fakeTimer.Reset(d)
}

func (t fakeTicker) Stop() {
	t.fakeTimer.Stop()
}

type fakeTimer struct {
	ch     chan time.Time
	clk    *FakeClock
	fn     func()
	when   int64
	period int64
}

func newFakeTimer(clk *FakeClock, when int64, period int64, fn func()) *fakeTimer {
	return &fakeTimer{
		ch:     make(chan time.Time, 1),
		clk:    clk,
		fn:     fn,
		when:   when,
		period: period,
	}
}

func (t *fakeTimer) C() <-chan time.Time {
	return t.ch
}

func (t *fakeTimer) Reset(d time.Duration) bool {
	return t.clk.resetTimer(t, int64(d)) > 0
}

func (t *fakeTimer) Stop() bool {
	return t.clk.stopTimer(t) > 0
}

func (t *fakeTimer) tick(now int64) {
	if t.fn != nil {
		go t.fn()
		return
	}

	ts := time.Unix(0, now)
	select {
	case t.ch <- ts:
	default:
		select {
		case <-t.ch:
		default:
		}

		select {
		case t.ch <- ts:
		default:
		}
	}
}
