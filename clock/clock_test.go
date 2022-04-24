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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/chrono/clock"
	"go.uber.org/atomic"
)

func TestNewClockErrors(t *testing.T) {
	cases := []struct {
		name      string
		opts      []clock.Option
		expectErr bool
	}{
		{
			name:      "no opts",
			opts:      nil,
			expectErr: false,
		},
		{
			name:      "default opts",
			opts:      []clock.Option{clock.DefaultOptions()},
			expectErr: false,
		},
		{
			name:      "with nanotime clock",
			opts:      []clock.Option{clock.WithNanotimeFunc(clock.DefaultNanotimeFunc())},
			expectErr: false,
		},
		{
			name:      "with wall clock",
			opts:      []clock.Option{clock.WithTimeFunc(clock.DefaultTimeFunc())},
			expectErr: false,
		},
		{
			name: "with both clocks",
			opts: []clock.Option{
				clock.WithTimeFunc(clock.DefaultTimeFunc()),
				clock.WithNanotimeFunc(clock.DefaultNanotimeFunc()),
			},
			expectErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			clk, err := clock.NewClock(tt.opts...)
			if tt.expectErr {
				require.Error(t, err)
				require.Nil(t, clk)
			} else {
				require.NoError(t, err)
				require.NotNil(t, clk)
			}
		})
	}
}

func TestNewClockResultingClock(t *testing.T) {
	var (
		nanotimeFunc = func() int64 { return 123 }
		timeFunc     = func() time.Time { return time.Unix(0, 456) }
		cases        = []struct {
			name           string
			opts           []clock.Option
			expectNanotime int64
		}{
			{
				name:           "with nanotime clock",
				opts:           []clock.Option{clock.WithNanotimeFunc(nanotimeFunc)},
				expectNanotime: nanotimeFunc(),
			},
			{
				name: "with nanotime clock via options",
				opts: []clock.Option{clock.Options{
					NanotimeFunc: nanotimeFunc,
				}},
				expectNanotime: nanotimeFunc(),
			},
			{
				name:           "with wall clock",
				opts:           []clock.Option{clock.WithTimeFunc(timeFunc)},
				expectNanotime: timeFunc().UnixNano(),
			},
			{
				name: "with both clocks",
				opts: []clock.Option{
					clock.WithTimeFunc(timeFunc),
					clock.WithNanotimeFunc(nanotimeFunc),
				},
				expectNanotime: nanotimeFunc(),
			},
		}
	)

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			clk := newTestClock(t, tt.opts...)
			require.Equal(t, tt.expectNanotime, clk.Nanotime())
		})
	}
}

func TestSpecializedClockConstructors(t *testing.T) {
	cases := []struct {
		name  string
		clock clock.Clock
	}{
		{
			name:  "monotonic",
			clock: clock.NewMonotonicClock(),
		},
		{
			name:  "wall",
			clock: clock.NewWallClock(),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cur := tt.clock.Nanotime()
			for i := 0; i < 100; i++ {
				time.Sleep(time.Millisecond)
				now := tt.clock.Nanotime()
				require.Greater(t, now, cur)
				cur = now
			}
		})
	}
}

func TestClockNewTimer(t *testing.T) {
	cases := []struct {
		name string
		opts []clock.Option
	}{
		{
			name: "nanotime",
			opts: []clock.Option{clock.WithNanotimeFunc(clock.DefaultNanotimeFunc())},
		},
		{
			name: "wall",
			opts: []clock.Option{clock.WithTimeFunc(clock.DefaultTimeFunc())},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				clk   = newTestClock(t, tt.opts...)
				timer = clk.NewTimer(time.Millisecond)
			)

			requireTick(t, timer.C())
			require.False(t, timer.Stop())
		})
	}
}

func TestClockNewTicker(t *testing.T) {
	cases := []struct {
		name string
		opts []clock.Option
	}{
		{
			name: "nanotime",
			opts: []clock.Option{clock.WithNanotimeFunc(clock.DefaultNanotimeFunc())},
		},
		{
			name: "wall",
			opts: []clock.Option{clock.WithTimeFunc(clock.DefaultTimeFunc())},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				clk   = newTestClock(t, tt.opts...)
				timer = clk.NewTicker(time.Millisecond)
			)
			defer timer.Stop()

			requireTick(t, timer.C())
		})
	}
}

func TestClockSince(t *testing.T) {
	cases := []struct {
		name string
		opts []clock.Option
	}{
		{
			name: "nanotime",
			opts: []clock.Option{clock.WithNanotimeFunc(clock.DefaultNanotimeFunc())},
		},
		{
			name: "wall",
			opts: []clock.Option{clock.WithTimeFunc(clock.DefaultTimeFunc())},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			clk := newTestClock(t, tt.opts...)
			require.InEpsilon(t, clk.Now().UnixNano(), clk.Since(time.Unix(0, 0)), float64(time.Second))
			require.InEpsilon(t, clk.Nanotime(), clk.SinceNanotime(0), float64(time.Second))
		})
	}
}

func TestClockAfter(t *testing.T) {
	cases := []struct {
		name string
		opts []clock.Option
	}{
		{
			name: "nanotime",
			opts: []clock.Option{clock.WithNanotimeFunc(clock.DefaultNanotimeFunc())},
		},
		{
			name: "wall",
			opts: []clock.Option{clock.WithTimeFunc(clock.DefaultTimeFunc())},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			clk := newTestClock(t, tt.opts...)
			requireTick(t, clk.After(time.Millisecond))

			var called atomic.Bool
			clk.AfterFunc(time.Millisecond, func() { called.Store(true) })
			waitFor(t, time.Second, called.Load)
		})
	}
}

func TestClockTick(t *testing.T) {
	cases := []struct {
		name string
		opts []clock.Option
	}{
		{
			name: "nanotime",
			opts: []clock.Option{clock.WithNanotimeFunc(clock.DefaultNanotimeFunc())},
		},
		{
			name: "wall",
			opts: []clock.Option{clock.WithTimeFunc(clock.DefaultTimeFunc())},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				clk     = newTestClock(t, tt.opts...)
				tickerC = clk.Tick(time.Millisecond)
			)

			for i := 0; i < 10; i++ {
				requireTick(t, tickerC)
			}
		})
	}
}

func TestClockSleep(t *testing.T) {
	cases := []struct {
		name string
		opts []clock.Option
	}{
		{
			name: "nanotime",
			opts: []clock.Option{clock.WithNanotimeFunc(clock.DefaultNanotimeFunc())},
		},
		{
			name: "wall",
			opts: []clock.Option{clock.WithTimeFunc(clock.DefaultTimeFunc())},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				clk       = newTestClock(t, tt.opts...)
				sleepdone = make(chan struct{})
			)

			go func() {
				defer close(sleepdone)
				clk.Sleep(50 * time.Millisecond)
			}()

			start := clk.Nanotime()
			select {
			case <-time.After(time.Second):
				require.Fail(t, "did not wake")
			case <-sleepdone:
				since := clk.SinceNanotime(start)
				require.InEpsilon(t, 50*time.Millisecond, since, float64(25*time.Millisecond))
			}
		})
	}
}

func newTestClock(t *testing.T, opts ...clock.Option) clock.Clock {
	clk, err := clock.NewClock(opts...)
	require.NoError(t, err)
	return clk
}
