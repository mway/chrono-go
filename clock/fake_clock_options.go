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

package clock

// A FakeHookFilter is a positive filter to match timers and/or tickers for
// hooks registered on a FakeClock.
type FakeHookFilter uint8

const (
	// FilterNone will match nothing.
	FilterNone FakeHookFilter = iota
	// FilterTimers will match timers.
	FilterTimers
	// FilterTickers will match tickers.
	FilterTickers
	// FilterAll will match both timers and tickers.
	FilterAll
)

// FakeOptions configure a FakeClock.
type FakeOptions struct {
	// Hooks are the FakeHooks to be registered.
	Hooks []FakeHook
}

// DefaultFakeOptions returns a new FakeOptions with default values.
func DefaultFakeOptions() FakeOptions {
	return FakeOptions{}
}

// With returns a new FakeOptions based on o with opts merged in.
func (o FakeOptions) With(opts ...FakeOption) FakeOptions {
	for _, opt := range opts {
		opt.apply(&o)
	}

	return o
}

func (o FakeOptions) apply(other *FakeOptions) {
	if len(o.Hooks) > 0 {
		other.Hooks = append(other.Hooks, o.Hooks...)
	}
}

// A FakeOption configures a FakeClock.
type FakeOption interface {
	apply(*FakeOptions)
}

// WithFakeHooks returns a FakeOption that appends the given hooks to a
// FakeClock.
func WithFakeHooks(hooks ...FakeHook) FakeOption {
	return fakeOptionFunc(func(o *FakeOptions) {
		o.Hooks = append(o.Hooks, hooks...)
	})
}

type fakeOptionFunc func(*FakeOptions)

func (f fakeOptionFunc) apply(o *FakeOptions) {
	f(o)
}
