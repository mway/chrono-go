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

package periodic

import (
	"go.mway.dev/chrono/clock"
)

var _defaultStartOptions = startOptions{
	Clock: clock.NewMonotonicClock(),
}

type startOptions struct {
	Clock clock.Clock
}

func defaultStartOptions() startOptions {
	return _defaultStartOptions
}

// With returns a new [StartOptions] with opts merged on top of o.
func (o startOptions) With(opts ...StartOption) startOptions {
	for _, opt := range opts {
		opt.apply(&o)
	}
	return o
}

// A StartOption is passed to [Start] to configure a [Handle].
type StartOption interface {
	apply(*startOptions)
}

// WithClock returns a [StartOption] that configures a [Handle] to use the
// given [clock.Clock] for measuring time.
func WithClock(clk clock.Clock) StartOption {
	return startOptionFunc(func(dst *startOptions) {
		if clk != nil {
			dst.Clock = clk
		}
	})
}

type startOptionFunc func(*startOptions)

func (f startOptionFunc) apply(dst *startOptions) {
	f(dst)
}
