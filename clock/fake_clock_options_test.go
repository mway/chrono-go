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

package clock_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/chrono/clock"
)

func TestFakeOptions(t *testing.T) {
	var (
		calls int
		base  = clock.FakeOptions{
			Hooks: []clock.FakeHook{
				{
					OnCreate: func(*clock.FakeClock, time.Duration) {
						calls++
					},
				},
			},
		}
		merged = base.With(clock.FakeOptions{
			Hooks: []clock.FakeHook{
				{
					OnCreate: func(*clock.FakeClock, time.Duration) {
						calls++
					},
				},
			},
		})
	)

	for _, hook := range merged.Hooks {
		hook.OnCreate(nil, 0)
	}

	require.Equal(t, 2, calls)
}
