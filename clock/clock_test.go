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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/chrono/clock"
	"go.uber.org/atomic"
)

var (
	_withNanotimeFunc = clock.WithNanotimeFunc(clock.DefaultNanotimeFunc())
	_withTimeFunc     = clock.WithTimeFunc(clock.DefaultTimeFunc())
)

func TestNewClock(t *testing.T) {
	cases := map[string]struct {
		opts      []clock.Option
		expectErr bool
	}{
		"no opts": {
			opts:      nil,
			expectErr: false,
		},
		"default opts": {
			opts:      []clock.Option{clock.DefaultOptions()},
			expectErr: false,
		},
		"with nanotime func": {
			opts:      []clock.Option{_withNanotimeFunc},
			expectErr: false,
		},
		"with time func": {
			opts:      []clock.Option{_withTimeFunc},
			expectErr: false,
		},
		"with nanotime and time funcs": {
			opts: []clock.Option{
				_withNanotimeFunc,
				_withTimeFunc,
			},
			expectErr: false,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
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

func TestNewClock_Funcs(t *testing.T) {
	var (
		nanotimeFunc = func() int64 { return 123 }
		timeFunc     = func() time.Time { return time.Unix(0, 456) }
		cases        = map[string]struct {
			opts           []clock.Option
			expectNanotime int64
		}{
			"with nanotime func": {
				opts:           []clock.Option{clock.WithNanotimeFunc(nanotimeFunc)},
				expectNanotime: nanotimeFunc(),
			},
			"with nanotime func options": {
				opts: []clock.Option{clock.Options{
					NanotimeFunc: nanotimeFunc,
				}},
				expectNanotime: nanotimeFunc(),
			},
			"with time func": {
				opts:           []clock.Option{clock.WithTimeFunc(timeFunc)},
				expectNanotime: timeFunc().UnixNano(),
			},
			"with time func options": {
				opts: []clock.Option{clock.Options{
					TimeFunc: timeFunc,
				}},
				expectNanotime: timeFunc().UnixNano(),
			},
			"with nanotime and time funcs": {
				opts: []clock.Option{
					clock.WithTimeFunc(timeFunc),
					clock.WithNanotimeFunc(nanotimeFunc),
				},
				expectNanotime: nanotimeFunc(),
			},
		}
	)

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			clk := newTestClock(t, tt.opts...)
			require.Equal(t, tt.expectNanotime, clk.Nanotime())
		})
	}
}

func TestMustClock(t *testing.T) {
	require.Panics(t, func() {
		clock.MustClock(nil, errors.New("error"))
	})

	require.NotPanics(t, func() {
		clk := clock.NewFakeClock()
		require.Equal(t, clk, clock.MustClock(clk, nil))
	})
}

func TestSpecializedClockConstructors(t *testing.T) {
	cases := map[string]struct {
		clock clock.Clock
	}{
		"NewMonotonicClock": {
			clock: clock.NewMonotonicClock(),
		},
		"NewWallClock": {
			clock: clock.NewWallClock(),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
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

func TestClock_NewTimer(t *testing.T) {
	cases := map[string]struct {
		opts []clock.Option
	}{
		"nanotime func": {
			opts: []clock.Option{_withNanotimeFunc},
		},
		"time func": {
			opts: []clock.Option{_withTimeFunc},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			var (
				clk   = newTestClock(t, tt.opts...)
				timer = clk.NewTimer(time.Millisecond)
			)

			requireTick(t, timer.C)
			require.False(t, timer.Stop())
			timer.Reset(time.Second)
			require.True(t, timer.Stop())
		})
	}
}

func TestClock_NewTicker(t *testing.T) {
	cases := map[string]struct {
		opts []clock.Option
	}{
		"nanotime func": {
			opts: []clock.Option{_withNanotimeFunc},
		},
		"time func": {
			opts: []clock.Option{_withTimeFunc},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			var (
				clk    = newTestClock(t, tt.opts...)
				ticker = clk.NewTicker(time.Millisecond)
			)
			defer ticker.Stop()

			requireTick(t, ticker.C)
			ticker.Stop()
			ticker.Reset(time.Millisecond)
			requireTick(t, ticker.C)
		})
	}
}

func TestClock_Since(t *testing.T) {
	cases := map[string]struct {
		opts []clock.Option
	}{
		"nanotime func": {
			opts: []clock.Option{_withNanotimeFunc},
		},
		"time func": {
			opts: []clock.Option{_withTimeFunc},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			clk := newTestClock(t, tt.opts...)
			require.InEpsilon(
				t,
				clk.Now().UnixNano(),
				clk.Since(time.Unix(0, 0)),
				float64(time.Second),
			)
			require.InEpsilon(
				t,
				clk.Nanotime(),
				clk.SinceNanotime(0),
				float64(time.Second),
			)
		})
	}
}

func TestClock_After(t *testing.T) {
	cases := map[string]struct {
		opts []clock.Option
	}{
		"nanotime func": {
			opts: []clock.Option{_withNanotimeFunc},
		},
		"time func": {
			opts: []clock.Option{_withTimeFunc},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			clk := newTestClock(t, tt.opts...)
			requireTick(t, clk.After(time.Millisecond))

			var called atomic.Bool
			clk.AfterFunc(time.Millisecond, func() { called.Store(true) })
			waitFor(t, time.Second, called.Load)
		})
	}
}

func TestClock_Tick(t *testing.T) {
	cases := map[string]struct {
		opts []clock.Option
	}{
		"nanotime func": {
			opts: []clock.Option{_withNanotimeFunc},
		},
		"time func": {
			opts: []clock.Option{_withTimeFunc},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
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

func TestClock_Sleep(t *testing.T) {
	cases := map[string]struct {
		name string
		opts []clock.Option
	}{
		"nanotime func": {
			opts: []clock.Option{_withNanotimeFunc},
		},
		"time func": {
			opts: []clock.Option{_withTimeFunc},
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
				require.InEpsilon(
					t,
					50*time.Millisecond,
					since,
					float64(25*time.Millisecond),
				)
			}
		})
	}
}

func TestClock_Stopwatch(t *testing.T) {
	var (
		cases = map[string]struct {
			giveClock clock.Clock
		}{
			"nanotime func": {
				giveClock: clock.NewMonotonicClock(),
			},
			"time func": {
				giveClock: clock.NewWallClock(),
			},
		}
		waitElapse = func(s *clock.Stopwatch, d time.Duration) time.Duration {
			for {
				if x := s.Elapsed(); x >= d {
					return d
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	)

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			var (
				stopwatch = tt.giveClock.NewStopwatch()
				elapsed   = waitElapse(stopwatch, 10*time.Millisecond)
			)

			require.GreaterOrEqual(t, elapsed, 10*time.Millisecond)
			require.GreaterOrEqual(t, stopwatch.Reset(), elapsed)
			require.Less(t, stopwatch.Elapsed(), 10*time.Millisecond)

			elapsed = waitElapse(stopwatch, 10*time.Millisecond)
			require.GreaterOrEqual(t, elapsed, 10*time.Millisecond)
			require.GreaterOrEqual(t, stopwatch.Reset(), elapsed)
		})
	}
}

func newTestClock(t *testing.T, opts ...clock.Option) clock.Clock {
	clk, err := clock.NewClock(opts...)
	require.NoError(t, err)
	return clk
}
