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
	"go.mway.dev/chrono/clock"
	"go.mway.dev/errors"
	"go.uber.org/multierr"
)

var (
	// ErrNilClock indicates that the given clock is nil.
	ErrNilClock = errors.New("nil clock provided")

	_defaultOptions = Options{
		Clock: clock.NewMonotonicClock(),
	}

	_ Option = Options{}
)

// Options configure a Stopwatch.
type Options struct {
	Clock clock.Clock
}

// DefaultOptions returns a new, default Options.
func DefaultOptions() Options {
	return _defaultOptions
}

// Validate returns an error if this Options contains invalid data.
func (o Options) Validate() error {
	var err error

	if o.Clock == nil {
		err = multierr.Append(err, ErrNilClock)
	}

	return errors.Wrap(err, "invalid Options")
}

// With returns a new Options based on o with the given opts merged onto it.
func (o Options) With(opts ...Option) Options {
	for _, opt := range opts {
		opt.apply(&o)
	}

	return o
}

func (o Options) apply(other *Options) {
	if o.Clock != nil {
		other.Clock = o.Clock
	}
}

// An Option configures a Stopwatch.
type Option interface {
	apply(*Options)
}

// WithClock returns an Option that configures a Stopwatch to use the given
// clock.
func WithClock(clk clock.Clock) Option {
	return optionFunc(func(o *Options) {
		o.Clock = clk
	})
}

type optionFunc func(*Options)

func (f optionFunc) apply(o *Options) {
	f(o)
}
