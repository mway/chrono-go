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

package stopwatch

import (
	"time"

	"go.mway.dev/chrono/clock"
	"go.uber.org/atomic"
)

// A Stopwatch measures elapsed time.
type Stopwatch struct {
	clock clock.Clock
	start atomic.Int64
}

// New creates a new Stopwatch with the given options.
func New(opts ...Option) (*Stopwatch, error) {
	options := DefaultOptions().With(opts...)
	if err := options.Validate(); err != nil {
		return nil, err
	}

	s := &Stopwatch{
		clock: options.Clock,
	}
	s.Reset()

	return s, nil
}

// Elapsed returns the time elapsed since the last call to Reset.
func (s *Stopwatch) Elapsed() time.Duration {
	return time.Duration(s.clock.Nanotime() - s.start.Load())
}

// Reset sets the Stopwatch's internal time to the current time.
func (s *Stopwatch) Reset() {
	s.start.Store(s.clock.Nanotime())
}
