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

package stopwatch_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/chrono/clock"
	"go.mway.dev/chrono/stopwatch"
)

func TestStopwatch(t *testing.T) {
	mockclock := clock.NewFakeClock()

	sw, err := stopwatch.New(stopwatch.WithClock(mockclock))
	require.NoError(t, err)
	require.Equal(t, 0*time.Second, sw.Elapsed())

	mockclock.SetNanotime(int64(time.Second))
	require.Equal(t, time.Second, sw.Elapsed())

	sw.Reset()
	require.Equal(t, 0*time.Second, sw.Elapsed())

	mockclock.SetNanotime(int64(3 * time.Second))
	require.Equal(t, 2*time.Second, sw.Elapsed())
}

func TestNew(t *testing.T) {
	cases := map[string]struct {
		opts      []stopwatch.Option
		expectErr error
	}{
		"no opts": {
			opts:      nil,
			expectErr: nil,
		},
		"valid clock": {
			opts: []stopwatch.Option{
				stopwatch.WithClock(clock.NewFakeClock()),
			},
			expectErr: nil,
		},
		"valid clock options": {
			opts: []stopwatch.Option{
				stopwatch.Options{
					Clock: clock.NewFakeClock(),
				},
			},
			expectErr: nil,
		},
		"nil clock": {
			opts: []stopwatch.Option{
				stopwatch.WithClock(nil),
			},
			expectErr: stopwatch.ErrNilClock,
		},
		"nil clock options": {
			opts: []stopwatch.Option{
				stopwatch.Options{
					Clock: nil,
				},
			},
			expectErr: nil,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := stopwatch.New(tt.opts...)
			require.ErrorIs(t, err, tt.expectErr)
		})
	}
}

func TestStopwatch_Elapsed(t *testing.T) {
	var (
		clk = clock.NewFakeClock()
		sw  = newStopwatch(t, stopwatch.WithClock(clk))
	)

	for i := 0; i < 1e4; i++ {
		require.EqualValues(t, 0, sw.Elapsed())
	}

	clk.Add(time.Second)
	require.EqualValues(t, time.Second, sw.Elapsed())
}

func newStopwatch(t *testing.T, opts ...stopwatch.Option) *stopwatch.Stopwatch {
	sw, err := stopwatch.New(opts...)
	require.NoError(t, err)
	return sw
}
