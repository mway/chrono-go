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

import (
	"time"

	"go.mway.dev/chrono"
)

// A TimeFunc is a function that returns time as a [time.Time] object.
type TimeFunc = func() time.Time

// DefaultTimeFunc returns a new [TimeFunc] that uses [time.Now] to tell time.
func DefaultTimeFunc() TimeFunc {
	return time.Now
}

// A NanotimeFunc is a function that returns time as integer nanoseconds.
type NanotimeFunc = func() int64

// DefaultNanotimeFunc returns a new [NanotimeFunc] that uses [chrono.Nanotime]
// to tell time.
func DefaultNanotimeFunc() NanotimeFunc {
	return chrono.Nanotime
}

// Options configure a [Clock].
type Options struct {
	// TimeFunc configures the [TimeFunc] for a [Clock].
	// If both TimeFunc and NanotimeFunc are provided, NanotimeFunc is used.
	TimeFunc TimeFunc
	// NanotimeFunc configures the [NanotimeFunc] for a [Clock].
	// If both TimeFunc and NanotimeFunc are provided, NanotimeFunc is used.
	NanotimeFunc NanotimeFunc
}

// DefaultOptions returns a new [Options] with sane defaults.
func DefaultOptions() Options {
	return Options{
		TimeFunc: DefaultTimeFunc(),
	}
}

func (o Options) apply(opts *Options) {
	if o.TimeFunc != nil {
		opts.TimeFunc = o.TimeFunc
	}

	if o.NanotimeFunc != nil {
		opts.NanotimeFunc = o.NanotimeFunc
	}
}

// An Option configures a Clock.
type Option interface {
	apply(*Options)
}

type optionFunc func(*Options)

func (f optionFunc) apply(o *Options) {
	f(o)
}

// WithNanotimeFunc returns an [Option] that configures a [Clock] to use f as
// its time function.
func WithNanotimeFunc(f NanotimeFunc) Option {
	return optionFunc(func(o *Options) {
		o.NanotimeFunc = f
		o.TimeFunc = nil
	})
}

// WithTimeFunc returns an [Option] that configures a [Clock] to use f as its
// time function.
func WithTimeFunc(f TimeFunc) Option {
	return optionFunc(func(o *Options) {
		o.TimeFunc = f
		o.NanotimeFunc = nil
	})
}
