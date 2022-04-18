package chrono_test

import (
	"testing"
	"time"

	"go.mway.dev/chrono"
)

func BenchmarkThrottledClock(b *testing.B) {
	var (
		cases = []struct {
			name  string
			nowfn chrono.NanoFunc
		}{
			{
				name:  "mono",
				nowfn: chrono.NewMonotonicNanoFunc(),
			},
			{
				name:  "wall",
				nowfn: chrono.NewWallNanoFunc(),
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
					clock := chrono.NewThrottledClock(tt.nowfn, dur)
					defer clock.Stop()

					b.ReportAllocs()
					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						nanos = clock.Nanos()
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
		clock := chrono.NewThrottledMonotonicClock(time.Microsecond)
		defer clock.Stop()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			nanos = clock.Nanos()
		}
	})

	b.Run("now", func(b *testing.B) {
		clock := chrono.NewThrottledMonotonicClock(time.Microsecond)
		defer clock.Stop()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			now = clock.Now()
		}
	})

	_ = nanos
	_ = now
}
