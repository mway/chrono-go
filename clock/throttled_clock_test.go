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

package clock_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/chrono/clock"
	"go.mway.dev/math"
	"go.uber.org/atomic"
)

func TestThrottledClock(t *testing.T) {
	cases := []struct {
		name    string
		clockFn func(time.Duration) *clock.ThrottledClock
	}{
		{
			name: "mono",
			clockFn: func(d time.Duration) *clock.ThrottledClock {
				return clock.NewThrottledMonotonicClock(d)
			},
		},
		{
			name: "wall",
			clockFn: func(d time.Duration) *clock.ThrottledClock {
				return clock.NewThrottledWallClock(d)
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				clock     = tt.clockFn(time.Millisecond)
				prevNanos = clock.Nanotime()
				prevTime  = clock.Now()
			)

			waitForChange(t, clock, prevNanos)

			require.True(t, clock.Nanotime() > prevNanos, "nanotime did not increase")
			require.True(t, clock.Now().After(prevTime), "time did not increase")
		})
	}
}

func TestThrottledClockInternals(t *testing.T) {
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
