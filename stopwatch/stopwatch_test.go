package stopwatch_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/chrono/clock"
	"go.mway.dev/chrono/stopwatch"
)

func TestStopwatch(t *testing.T) {
	mockclock := clock.NewFakeClock()

	sw, err := stopwatch.New(stopwatch.WithClock(mockclock))
	require.NoError(t, err)
	require.Equal(t, 0*time.Second, sw.Elapsed())

	mockclock.SetNanotime(int64(time.Second))
	require.Equal(t, time.Second, sw.Elapsed())

	sw.Reset()
	require.Equal(t, 0*time.Second, sw.Elapsed())

	mockclock.SetNanotime(int64(3 * time.Second))
	require.Equal(t, 2*time.Second, sw.Elapsed())
}

func TestNew(t *testing.T) {
	cases := map[string]struct {
		opts      []stopwatch.Option
		expectErr error
	}{
		"no opts": {
			opts:      nil,
			expectErr: nil,
		},
		"valid clock": {
			opts: []stopwatch.Option{
				stopwatch.WithClock(clock.NewFakeClock()),
			},
			expectErr: nil,
		},
		"valid clock options": {
			opts: []stopwatch.Option{
				stopwatch.Options{
					Clock: clock.NewFakeClock(),
				},
			},
			expectErr: nil,
		},
		"nil clock": {
			opts: []stopwatch.Option{
				stopwatch.WithClock(nil),
			},
			expectErr: stopwatch.ErrNilClock,
		},
		"nil clock options": {
			opts: []stopwatch.Option{
				stopwatch.Options{
					Clock: nil,
				},
			},
			expectErr: nil,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := stopwatch.New(tt.opts...)
			require.ErrorIs(t, err, tt.expectErr)
		})
	}
}
