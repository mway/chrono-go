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
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/chrono/clock"
	"go.mway.dev/math"
	"go.uber.org/atomic"
)

func TestThrottledClock_Constructors(t *testing.T) {
	cases := map[string]struct {
		clockFn func(time.Duration) *clock.ThrottledClock
	}{
		"NewThrottledMonotonicClock": {
			clockFn: func(d time.Duration) *clock.ThrottledClock {
				return clock.NewThrottledMonotonicClock(d)
			},
		},
		"NewThrottledWallClock": {
			clockFn: func(d time.Duration) *clock.ThrottledClock {
				return clock.NewThrottledWallClock(d)
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			var (
				clk       = tt.clockFn(time.Millisecond)
				prevNanos = clk.Nanotime()
				prevTime  = clk.Now()
			)
			defer clk.Stop()

			waitForChange(t, clk, prevNanos)

			require.True(
				t,
				clk.Nanotime() > prevNanos,
				"nanotime did not increase",
			)
			require.True(
				t,
				clk.Now().After(prevTime),
				"time did not increase",
			)
		})
	}
}

func TestNewThrottledClock_Panic(t *testing.T) {
	require.Panics(t, func() {
		clock.NewThrottledClock(nil, -1)
	})
	require.Panics(t, func() {
		clock.NewThrottledClock(func() int64 { return 0 }, -1)
	})
}

//nolint:gocyclo
func TestThrottledClock_Timers(t *testing.T) {
	clk := clock.NewThrottledClock(func() int64 { return 0 }, time.Minute)
	defer clk.Stop()

	var (
		first = clk.Nanotime()
		wg    sync.WaitGroup
	)

	// ThrottledClock.NewTimer
	wg.Add(1)
	go func() {
		defer wg.Done()

		timer := clk.NewTimer(time.Millisecond)
		defer timer.Stop()

		select {
		case _, ok := <-timer.C:
			require.True(t, ok, "zero time returned")
		case <-time.After(time.Second):
			require.FailNow(t, "timer did not fire")
		}
	}()

	// ThrottledClock.NewTicker
	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := clk.NewTicker(time.Millisecond)
		defer ticker.Stop()

		for i := 0; i < 10; i++ {
			select {
			case <-ticker.C:
			case <-time.After(time.Second):
				require.FailNow(t, "timer did not fire")
			}
		}
	}()

	// ThrottledClock.Tick
	wg.Add(1)
	go func() {
		defer wg.Done()

		tickerC := clk.Tick(time.Millisecond)

		for i := 0; i < 10; i++ {
			select {
			case <-tickerC:
			case <-time.After(time.Second):
				require.FailNow(t, "timer did not fire")
			}
		}
	}()

	// ThrottledClock.After
	wg.Add(1)
	go func() {
		defer wg.Done()

		timerC := clk.After(time.Millisecond)

		select {
		case <-timerC:
		case <-time.After(time.Second):
			require.FailNow(t, "timer did not fire")
		}
	}()

	// ThrottledClock.AfterFunc
	wg.Add(1)
	go func() {
		defer wg.Done()

		timerC := make(chan struct{})
		clk.AfterFunc(time.Millisecond, func() {
			close(timerC)
		})

		select {
		case <-timerC:
		case <-time.After(time.Second):
			require.FailNow(t, "timer did not fire")
		}
	}()

	// ThrottledClock.Sleep
	wg.Add(1)
	go func() {
		defer wg.Done()

		timerC := make(chan struct{})
		go func() {
			clk.Sleep(10 * time.Millisecond)
			close(timerC)
		}()

		select {
		case <-timerC:
		case <-time.After(time.Second):
			require.FailNow(t, "timer did not fire")
		}
	}()

	wg.Wait()
	require.Equal(t, first, clk.Nanotime())
}

func TestThrottledClock_Since(t *testing.T) {
	clk := clock.NewThrottledClock(func() int64 { return 123 }, time.Minute)
	defer clk.Stop()

	require.Equal(t, 23*time.Nanosecond, clk.Since(time.Unix(0, 100)))
	require.Equal(t, 23*time.Nanosecond, clk.SinceNanotime(100))
}

func TestThrottledClock_Internals(t *testing.T) {
	var (
		now   = atomic.NewInt64(123)
		nowfn = func() int64 {
			return now.Load()
		}
	)

	clk := clock.NewThrottledClock(nowfn, time.Microsecond)
	defer clk.Stop()

	require.Equal(t, now.Load(), clk.Nanotime())
	require.True(t, clk.Now().Equal(time.Unix(0, now.Load())))

	prev := now.Load()
	now.Store(456)
	waitForChange(t, clk, prev)

	require.Equal(t, now.Load(), clk.Nanotime())
	require.True(t, clk.Now().Equal(time.Unix(0, now.Load())))

	prev = now.Load()
	now.Store(1)
	waitForChange(t, clk, prev)

	require.Equal(t, now.Load(), clk.Nanotime())
	require.True(t, clk.Now().Equal(time.Unix(0, now.Load())))

	clk.Stop()

	prev = now.Load()
	now.Store(1)

	// The clock should no longer update once it is stopped.
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, prev, clk.Nanotime())
}

func TestThrottledClock_Stopwatch(t *testing.T) {
	var (
		now   = atomic.NewInt64(0)
		nowfn = func() int64 {
			return now.Load()
		}
	)

	clk := clock.NewThrottledClock(nowfn, time.Microsecond)
	defer clk.Stop()

	stopwatch := clk.NewStopwatch()
	require.Equal(t, 0*time.Second, stopwatch.Elapsed())

	now.Add(int64(time.Second))
	waitForChange(t, clk, 0)
	require.Equal(t, time.Second, stopwatch.Elapsed())

	now.Add(int64(time.Second))
	waitForChange(t, clk, int64(time.Second))
	require.Equal(t, 2*time.Second, stopwatch.Elapsed())
	require.Equal(t, 2*time.Second, stopwatch.Reset())
	require.Equal(t, 0*time.Second, stopwatch.Elapsed())

	now.Add(int64(time.Second))
	waitForChange(t, clk, int64(2*time.Second))
	require.Equal(t, time.Second, stopwatch.Elapsed())
}

func waitForChange(t *testing.T, clk *clock.ThrottledClock, prev int64) {
	var (
		done = make(chan struct{})
		stop atomic.Bool
	)

	go func() {
		defer close(done)

		for !stop.Load() && clk.Nanotime() == prev {
			time.Sleep(clk.Interval() / 2)
			runtime.Gosched()
		}
	}()

	wait := math.Max(time.Second, 2*clk.Interval())
	select {
	case <-time.After(wait):
		stop.Store(true)
	case <-done:
	}

	<-done
	require.NotEqual(t, prev, clk.Nanotime(), "clock did not update")
}
