package chrono_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/chrono"
	"go.mway.dev/math"
	"go.uber.org/atomic"
)

func TestThrottledClock(t *testing.T) {
	var (
		now   = atomic.NewInt64(123)
		nowfn = func() int64 {
			return now.Load()
		}
	)

	clock := chrono.NewThrottledClock(nowfn, time.Microsecond)
	defer clock.Stop()

	require.Equal(t, now.Load(), clock.Nanos())
	require.True(t, clock.Now().Equal(time.Unix(0, now.Load())))

	prev := now.Load()
	now.Store(456)
	waitForChange(t, clock, prev)

	require.Equal(t, now.Load(), clock.Nanos())
	require.True(t, clock.Now().Equal(time.Unix(0, now.Load())))

	prev = now.Load()
	now.Store(1)
	waitForChange(t, clock, prev)

	require.Equal(t, now.Load(), clock.Nanos())
	require.True(t, clock.Now().Equal(time.Unix(0, now.Load())))

	clock.Stop()

	prev = now.Load()
	now.Store(1)

	// The clock should no longer update once it is stopped.
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, prev, clock.Nanos())
}

func waitForChange(t *testing.T, clock *chrono.ThrottledClock, prev int64) {
	var (
		done = make(chan struct{})
		stop atomic.Bool
	)

	go func() {
		defer close(done)

		for !stop.Load() && clock.Nanos() == prev {
			time.Sleep(clock.Interval() / 2)
			runtime.Gosched()
		}
	}()

	wait := math.Max(time.Second, 2*clock.Interval())
	select {
	case <-time.After(wait):
		stop.Store(true)
	case <-done:
	}

	<-done
	require.NotEqual(t, prev, clock.Nanos(), "clock did not update")
}
