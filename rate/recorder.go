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

package rate

import (
	"time"

	"go.mway.dev/chrono/clock"
	"go.uber.org/atomic"
)

// A Recorder records added counts and reports the rate of the total count over
// the elapsed time.
type Recorder struct {
	clock clock.Clock
	count atomic.Int64
	epoch atomic.Int64
}

// NewRecorder creates a new [Recorder] that uses the system's monotonic clock.
func NewRecorder() *Recorder {
	return NewRecorderWithClock(clock.NewMonotonicClock())
}

// NewRecorderWithClock returns a new [Recorder] that uses the given clock.
func NewRecorderWithClock(clk clock.Clock) *Recorder {
	r := &Recorder{
		clock: clk,
	}
	r.Reset()
	return r
}

// Add adds n to the running count.
func (r *Recorder) Add(n int) {
	r.count.Add(int64(n))
}

// Rate returns a [Rate] that represents the running count and time elapsed
// since the recorder's clock started.
func (r *Recorder) Rate() Rate {
	return Rate{
		count:   r.count.Load(),
		elapsed: r.clock.SinceNanotime(r.epoch.Load()),
	}
}

// Reset returns the current [Rate] and resets the recorder's running count and
// epoch.
func (r *Recorder) Reset() Rate {
	var (
		now     = r.clock.Nanotime()
		elapsed = time.Duration(now - r.epoch.Swap(now))
	)
	return Rate{
		count:   r.count.Swap(0),
		elapsed: elapsed,
	}
}

// A Rate is a count over a period of time.
type Rate struct {
	count   int64
	elapsed time.Duration
}

// Per returns the rate's count over the given period of time.
func (r Rate) Per(d time.Duration) float64 {
	return (float64(r.count) / float64(r.elapsed)) * float64(d)
}
