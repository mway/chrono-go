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

// Package clock provides clock-related types and utilities.
package clock

import (
	"errors"
	"time"
)

// ErrNoClockFunc is returned when creating a new Clock without a valid clock
// function.
var ErrNoClockFunc = errors.New("no clock function provided")

// A Clock tells time.
type Clock interface {
	After(time.Duration) <-chan time.Time
	AfterFunc(time.Duration, func()) Timer
	Now() time.Time
	Nanotime() int64
	Since(time.Time) time.Duration
	SinceNanotime(int64) time.Duration
	NewTimer(time.Duration) Timer
	NewTicker(time.Duration) Ticker
	Tick(time.Duration) <-chan time.Time
	Sleep(time.Duration)
}

// NewClock returns a new Clock based on the given options.
func NewClock(opts ...Option) (Clock, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt.apply(&options)
	}

	if options.NanotimeFunc != nil {
		return newMonotonicClock(options.NanotimeFunc), nil
	}

	return newWallClock(options.TimeFunc), nil
}

// MustClock panics if the given error is not nil, otherwise it returns the
// given clock.
func MustClock(clock Clock, err error) Clock {
	if err != nil {
		panic(err)
	}

	return clock
}

// NewMonotonicClock returns a new monotonic Clock.
func NewMonotonicClock() Clock {
	return MustClock(NewClock(WithNanotimeFunc(DefaultNanotimeFunc())))
}

// NewWallClock returns a new wall Clock.
func NewWallClock() Clock {
	return MustClock(NewClock(WithTimeFunc(DefaultTimeFunc())))
}
