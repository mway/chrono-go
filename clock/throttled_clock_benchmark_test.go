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

	"go.mway.dev/chrono/clock"
)

func BenchmarkThrottledClock(b *testing.B) {
	var (
		cases = []struct {
			name  string
			nowfn clock.NanoFunc
		}{
			{
				name:  "mono",
				nowfn: clock.NewMonotonicNanoFunc(),
			},
			{
				name:  "wall",
				nowfn: clock.NewWallNanoFunc(),
			},
		}
		intervals = []time.Duration{
			time.Second,
			100 * time.Millisecond,
			10 * time.Millisecond,
			time.Millisecond,
			100 * time.Microsecond,
			10 * time.Microsecond,
			time.Microsecond,
		}
		nanos int64
	)

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			for _, dur := range intervals {
				b.Run(dur.String(), func(b *testing.B) {
					clk := clock.NewThrottledClock(tt.nowfn, dur)
					defer clk.Stop()

					b.ReportAllocs()
					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						nanos = clk.Nanotime()
					}
				})
			}
		})
	}

	_ = nanos
}

func BenchmarkThrottledClockSources(b *testing.B) {
	var (
		nanos int64
		now   time.Time
	)

	b.Run("nanos", func(b *testing.B) {
		clk := clock.NewThrottledMonotonicClock(time.Microsecond)
		defer clk.Stop()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			nanos = clk.Nanotime()
		}
	})

	b.Run("now", func(b *testing.B) {
		clk := clock.NewThrottledMonotonicClock(time.Microsecond)
		defer clk.Stop()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			now = clk.Now()
		}
	})

	_ = nanos
	_ = now
}
