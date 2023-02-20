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

package clock

import "time"

// A Stopwatch measures elapsed time. A Stopwatch is created by calling
// [Clock.NewStopwatch].
type Stopwatch struct {
	clock Clock
	epoch int64
}

func newStopwatch(clk Clock) *Stopwatch {
	return &Stopwatch{
		clock: clk,
		epoch: clk.Nanotime(),
	}
}

// Elapsed returns the time elapsed since the last call to [Stopwatch.Reset].
func (s *Stopwatch) Elapsed() time.Duration {
	return time.Duration(s.clock.Nanotime() - s.epoch)
}

// Reset resets the stopwatch to zero, returning the elapsed time since the
// last call to Reset.
func (s *Stopwatch) Reset() time.Duration {
	var (
		now     = s.clock.Nanotime()
		elapsed = time.Duration(now - s.epoch)
	)

	s.epoch = now
	return elapsed
}
