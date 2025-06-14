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

package clock_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/chrono"
	"go.mway.dev/chrono/clock"
	"go.uber.org/atomic"
)

func TestFakeClock_Add(t *testing.T) {
	clk := clock.NewFakeClock()
	requireClockIs(t, 0, clk)

	// Test moving forward.
	for i := int64(1); i <= 1000; i++ {
		prev := clk.Nanotime()
		clk.Add(time.Duration(i))
		requireClockIs(t, prev+i, clk)
	}

	// Test moving backward.
	for i := int64(1); i <= 1000; i++ {
		prev := clk.Nanotime()
		clk.Add(-time.Duration(i))
		requireClockIs(t, prev-i, clk)
	}

	// We should be back where we started.
	requireClockIs(t, 0, clk)
}

func TestFakeClock_SetTime(t *testing.T) {
	clk := clock.NewFakeClock()

	for i := int64(1); i <= 1000; i++ {
		clk.SetTime(time.Unix(0, i))
		requireClockIs(t, i, clk)
	}

	for i := int64(1000); i > 0; i-- {
		clk.SetTime(time.Unix(0, i))
		requireClockIs(t, i, clk)
	}
}

func TestFakeClock_SetNanotime(t *testing.T) {
	clk := clock.NewFakeClock()

	for i := int64(1); i <= 1000; i++ {
		clk.SetNanotime(i)
		requireClockIs(t, i, clk)
	}

	for i := int64(1000); i > 0; i-- {
		clk.SetNanotime(i)
		requireClockIs(t, i, clk)
	}
}

func TestFakeClock_After(t *testing.T) {
	var (
		clk    = clock.NewFakeClock()
		timerC = clk.After(time.Second)
	)

	require.NotNil(t, timerC)
	requireNoTick(t, timerC)

	clk.Add(time.Second)

	ts := requireTick(t, timerC)
	requireTimeIs(t, int64(time.Second), ts)

	// Start a new timer. This time, we want to advance the clock once to its
	// expiration, and then once again immediately after, to ensure that the
	// timer's channel only contains the first tick.
	timerC = clk.After(time.Second)
	clk.Add(time.Second)
	clk.Add(time.Second)

	ts = requireTick(t, timerC)
	requireTimeIs(t, 2*int64(time.Second), ts)
}

func TestFakeClock_AfterFunc(t *testing.T) {
	var (
		clk   = clock.NewFakeClock()
		calls = atomic.NewInt64(0)
		fn    = func() { calls.Inc() }
		timer = clk.AfterFunc(time.Second, fn)
	)

	for i := int64(0); i < 10; i++ {
		timer.Reset(time.Second)
		clk.Add(time.Second)
		waitFor(t, time.Second, func() bool {
			return calls.Load() == i+1
		})
	}
}

func TestFakeClockSince(t *testing.T) {
	var (
		clk   = clock.NewFakeClock()
		since int64
	)

	for i := int64(1); i < 1000; i++ {
		clk.SetNanotime(i)
		requireClockIs(t, i, clk)
		requireClockSince(t, i, since, clk)
	}
}

func TestFakeClock_NewTimer(t *testing.T) {
	var (
		clk   = clock.NewFakeClock()
		timer = clk.NewTimer(time.Second)
	)

	for i := int64(0); i < 10; i++ {
		requireNoTick(t, timer.C)
		clk.Add(time.Second)
		requireTick(t, timer.C)
		require.False(t, timer.Reset(time.Second))
	}

	// Cause the timer's tick channel to fill, and then tick again.
	clk.Add(time.Second)

	// Because the last tick wasn't consumed, the reported timestamp will be
	// in the past.
	want := clk.Nanotime()
	timer.Reset(time.Second)
	clk.Add(time.Second)
	ts := requireTick(t, timer.C)
	requireTimeIs(t, want, ts)

	require.False(t, timer.Stop())
	requireNoTick(t, timer.C)
}

func TestFakeClock_Timer_Zeroes(t *testing.T) {
	var (
		clk   = clock.NewFakeClock()
		timer *clock.Timer
	)

	require.NotPanics(t, func() {
		timer = clk.NewTimer(-1)
	})

	require.NotPanics(t, func() {
		timer = clk.NewTimer(0)
	})

	requireNoTick(t, timer.C)

	clk.Add(time.Second)
	requireTick(t, timer.C)

	// Ensure that resetting the timer will not panic if given a duration <= 0.
	require.NotPanics(t, func() {
		timer.Reset(-1)
	})
	require.NotPanics(t, func() {
		timer.Reset(0)
	})

	// The timer will still report that it has been stopped, because the clock
	// has not been changed.
	require.True(t, timer.Stop())
}

func TestFakeClock_NewTicker(t *testing.T) {
	var (
		clk    = clock.NewFakeClock()
		ticker = clk.NewTicker(time.Second)
	)

	for i := int64(0); i < 10; i++ {
		requireNoTick(t, ticker.C)
		clk.Add(time.Second)
		requireTick(t, ticker.C)

		if i%2 == 0 {
			ticker.Reset(time.Second)
		}
	}

	ticker.Stop()
	requireNoTick(t, ticker.C)

	require.Panics(t, func() {
		clk.NewTicker(-1)
	})

	require.Panics(t, func() {
		clk.NewTicker(0)
	})
}

//nolint:gocyclo
func TestFakeClock_Ticker_Goroutine(t *testing.T) {
	var (
		clk   = clock.NewFakeClock()
		ticks = make(chan struct{}, 10)
		ready = make(chan struct{})
		wg    sync.WaitGroup
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < cap(ticks); i++ {
		wg.Add(1)
		go func() {
			ticker := clk.NewTicker(time.Second)
			defer ticker.Stop()

			ready <- struct{}{}

			for {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return
				}

				select {
				case ticks <- struct{}{}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	maxWait := time.NewTicker(5 * time.Second)
	defer maxWait.Stop()

	for i := 0; i < cap(ticks); i++ {
		select {
		case <-ready:
		case <-maxWait.C:
			require.FailNow(t, "timed out waiting for ticker ticks")
		}
	}

	for i := 0; i < cap(ticks); i++ {
		clk.Add(time.Second)

		select {
		case <-ticks:
		case <-maxWait.C:
			require.FailNow(t, "timed out waiting for ticker ticks")
		}
	}
}

func TestFakeClock_Ticker_Zeroes(t *testing.T) {
	var (
		clk    = clock.NewFakeClock()
		ticker = clk.NewTicker(time.Second)
	)

	requireNoTick(t, ticker.C)

	clk.Add(time.Second)
	requireTick(t, ticker.C)

	// Ensure that the ticker will panic if given a duration <= 0.
	require.Panics(t, func() {
		ticker.Reset(-1)
	})
	require.Panics(t, func() {
		ticker.Reset(0)
	})
}

func TestFakeClock_Tick(t *testing.T) {
	var (
		clk     = clock.NewFakeClock()
		tickerC = clk.Tick(time.Second)
	)

	for i := int64(0); i < 10; i++ {
		requireNoTick(t, tickerC)
		clk.Add(time.Second)
		requireTick(t, tickerC)
	}

	require.Panics(t, func() {
		clk.Tick(-1)
	})

	require.Panics(t, func() {
		clk.Tick(0)
	})
}

func TestFakeClock_Sleep(t *testing.T) {
	var (
		clk       = clock.NewFakeClock()
		sleepdone = make(chan struct{})
	)
	go func() {
		defer close(sleepdone)
		clk.Sleep(time.Second)
	}()

	for range 10 {
		clk.Add(time.Second)
		time.Sleep(time.Millisecond)
	}

	select {
	case <-sleepdone:
	case <-time.After(time.Second):
		require.Fail(t, "sleep did not wake")
	}
}

func TestFakeClock_InterleavedTimers(t *testing.T) {
	var (
		clk    = clock.NewFakeClock()
		timer2 = clk.NewTimer(3 * time.Second)
		timer3 = clk.NewTimer(1 * time.Second)
		timer1 = clk.NewTimer(2 * time.Second)
	)

	timer2.Reset(2 * time.Second)
	timer3.Reset(3 * time.Second)
	timer1.Reset(1 * time.Second)

	// n.b. Handle channels in reverse order. We defined them above in order of
	//      descending duration, but since ticking is synchonous, evaluate them
	//      in sorted (ascending) order to ensure that the internals are doing
	//      what we expect them to.
	for i := int64(0); i < 10; i++ {
		requireNoTick(t, timer1.C)
		requireNoTick(t, timer2.C)
		requireNoTick(t, timer3.C)

		clk.Add(time.Second)
		requireTick(t, timer1.C)
		requireNoTick(t, timer2.C)
		requireNoTick(t, timer3.C)

		clk.Add(time.Second)
		requireNoTick(t, timer1.C)
		requireTick(t, timer2.C)
		requireNoTick(t, timer3.C)

		clk.Add(time.Second)
		requireNoTick(t, timer1.C)
		requireNoTick(t, timer2.C)
		requireTick(t, timer3.C)

		require.False(t, timer1.Reset(1*time.Second))
		require.False(t, timer2.Reset(2*time.Second))
		require.False(t, timer3.Reset(3*time.Second))
	}

	// Ensure that all timers covered by the time change fire.
	clk.Add(3 * time.Second)
	requireTick(t, timer1.C)
	requireTick(t, timer2.C)
	requireTick(t, timer3.C)

	require.False(t, timer1.Stop())
	require.False(t, timer2.Stop())
	require.False(t, timer3.Stop())

	requireNoTick(t, timer1.C)
	requireNoTick(t, timer2.C)
	requireNoTick(t, timer3.C)
}

func TestFakeClock_ManyTimers(t *testing.T) {
	var (
		clk    = clock.NewFakeClock()
		timers = make([]*clock.Timer, 0, 5)
	)

	for range cap(timers) {
		timers = append(timers, clk.NewTimer(time.Second))
	}

	for i := int64(0); i < 10; i++ {
		for j := range timers {
			requireNoTick(t, timers[j].C)
		}

		clk.Add(time.Second)

		for j := range timers {
			requireTick(t, timers[j].C)
			require.False(t, timers[j].Reset(time.Second))
		}
	}

	for j := range timers {
		require.True(t, timers[j].Stop())
		requireNoTick(t, timers[j].C)
	}
}

func TestFakeClock_Timer_DoubleStop(t *testing.T) {
	cases := [][]time.Duration{
		{time.Second, 5 * time.Second},
		{5 * time.Second, time.Second},
		{time.Second, time.Second},
		{time.Second, 2 * time.Second, 3 * time.Second},
		{3 * time.Second, 2 * time.Second, time.Second},
	}

	var (
		clk    = clock.NewFakeClock()
		timers []*clock.Timer
	)

	for _, durations := range cases {
		for _, dur := range durations {
			timers = append(timers, clk.NewTimer(dur))
		}

		for _, timer := range timers {
			require.True(t, timer.Stop())
			require.False(t, timer.Stop())
		}

		timers = timers[:0]
	}
}

func TestFakeClock_Stopwatch(t *testing.T) {
	var (
		clk       = clock.NewFakeClock()
		stopwatch = clk.NewStopwatch()
	)

	require.Equal(t, 0*time.Second, stopwatch.Elapsed())

	clk.Add(time.Second)
	require.Equal(t, time.Second, stopwatch.Elapsed())

	clk.Add(time.Second)
	require.Equal(t, 2*time.Second, stopwatch.Elapsed())
	require.Equal(t, 2*time.Second, stopwatch.Reset())
	require.Equal(t, 0*time.Second, stopwatch.Elapsed())

	clk.Add(time.Second)
	require.Equal(t, time.Second, stopwatch.Elapsed())
}

func requireClockSince(
	t *testing.T,
	expect int64,
	since int64,
	clk *clock.FakeClock,
) {
	require.EqualValues(t, expect, clk.Since(time.Unix(0, since)))
	require.EqualValues(t, expect, clk.SinceNanotime(since))
}

func requireClockIs(t *testing.T, expect int64, clk *clock.FakeClock) {
	t.Helper()
	requireTimeIs(t, expect, clk.Now())
	requireNanotimeIs(t, expect, clk.Nanotime())
}

func requireNanotimeIs(t *testing.T, expect int64, ns int64) {
	t.Helper()
	require.Equal(t, expect, ns)
}

func requireTimeIs(t *testing.T, expect int64, ts time.Time) {
	t.Helper()
	require.EqualValues(t, expect, ts.UnixNano())
}

func requireTick(t *testing.T, ch <-chan time.Time) (ts time.Time) {
	t.Helper()
	select {
	case ts = <-ch:
	case <-time.After(time.Second):
		require.Fail(t, "timed out waiting for tick")
	}
	return
}

func requireNoTick(t *testing.T, ch <-chan time.Time) {
	t.Helper()

	select {
	case <-ch:
		require.Fail(t, "unexpected tick")
	default:
	}
}

func waitFor(t *testing.T, d time.Duration, f func() bool) {
	t.Helper()

	start := chrono.Nanotime()
	for !f() {
		if time.Duration(chrono.Nanotime()-start) >= d {
			require.Fail(t, "timeout", "waited for %v", d)
		}
		time.Sleep(d >> 8)
	}
}
