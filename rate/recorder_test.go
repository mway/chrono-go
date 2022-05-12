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

package rate_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/chrono/clock"
	"go.mway.dev/chrono/rate"
)

func TestRecorder(t *testing.T) {
	var (
		clk      = clock.NewFakeClock()
		recorder = rate.NewRecorderWithClock(clk)
	)

	// A basic rate of 1M/s.
	clk.Add(time.Second)
	recorder.Add(1_000_000)
	rate := recorder.Rate()
	require.EqualValues(t, 1_000_000, rate.Per(time.Second))

	// Add another million without updating the clock, for a new rate of 2M/s.
	recorder.Add(1_000_000)
	rate = recorder.Rate()
	require.EqualValues(t, 2_000_000, rate.Per(time.Second))
	require.EqualValues(t, 4_000_000, rate.Per(2*time.Second))
	require.EqualValues(t, 1_000_000, rate.Per(500*time.Millisecond))

	// Add another second to the clock, doubling the amount of time that has
	// elapsed; the rate should drop by half.
	clk.Add(time.Second)
	rate = recorder.Rate()
	require.EqualValues(t, 1_000_000, rate.Per(time.Second))
	require.EqualValues(t, 2_000_000, rate.Per(2*time.Second))
	require.EqualValues(t, 10_000_000, rate.Per(10*time.Second))

	// Reset the timer to ensure that it starts fresh.
	recorder.Reset()
	recorder.Add(1_000_000)
	clk.Add(1000 * time.Second)
	rate = recorder.TakeRate()
	require.EqualValues(t, 1_000, rate.Per(time.Second))

	// Get a new rate after having called TakeRate to ensure that it's fresh.
	recorder.Add(1_000_000)
	clk.Add(time.Second)
	rate = recorder.Rate()
	require.EqualValues(t, 1_000_000, rate.Per(time.Second))
}

func TestRecorderRealTime(t *testing.T) {
	recorder := rate.NewRecorder()
	recorder.Add(1_000_000)

	rate := recorder.TakeRate()
	require.True(t, rate.Per(time.Nanosecond) > 100.0)
}
