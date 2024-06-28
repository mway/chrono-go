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

package periodic_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/chrono/clock"
	"go.mway.dev/chrono/periodic"
	"go.mway.dev/x/channels"
)

func TestStart(t *testing.T) {
	var (
		calls  = make(chan struct{})
		clk    = clock.NewFakeClock()
		handle = periodic.Start(
			time.Second,
			func(ctx context.Context) {
				select {
				case <-ctx.Done():
				case calls <- struct{}{}:
				}
			},
			periodic.WithClock(clk),
		)
	)

	defer handle.Stop()

	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	for seen := 0; seen < 1000; /* noincr */ {
		select {
		case <-calls:
			seen++
		case <-timeout.C:
			require.FailNow(t, "timed out waiting for periodic calls")
		default:
			clk.Add(time.Second)
		}
	}
}

func TestStart_Freespin(t *testing.T) {
	var (
		calls  = make(chan struct{})
		clk    = clock.NewFakeClock()
		handle = periodic.Start(
			-1,
			func(ctx context.Context) {
				select {
				case <-ctx.Done():
				case calls <- struct{}{}:
				}
			},
			periodic.WithClock(clk),
		)
	)

	defer handle.Stop()

	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	for seen := 0; seen < 1000; /* noincr */ {
		select {
		case <-calls:
			seen++
		case <-timeout.C:
			require.FailNow(t, "timed out waiting for periodic calls")
		default:
			clk.Add(time.Second)
		}
	}
}

func TestStartWithContext(t *testing.T) {
	var (
		calls  = make(chan struct{})
		clk    = clock.NewFakeClock()
		handle = periodic.StartWithContext(
			context.Background(),
			time.Second,
			func(ctx context.Context) {
				select {
				case <-ctx.Done():
				case calls <- struct{}{}:
				}
			},
			periodic.WithClock(clk),
		)
	)

	defer handle.Stop()

	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	for seen := 0; seen < 1000; /* noincr */ {
		select {
		case <-calls:
			seen++
		case <-timeout.C:
			require.FailNow(t, "timed out waiting for periodic calls")
		default:
			clk.Add(time.Second)
		}
	}
}

func TestStartWithContext_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var (
		calls  = make(chan struct{}, 1)
		clk    = clock.NewFakeClock()
		handle = periodic.StartWithContext(
			ctx,
			time.Second,
			func(ctx context.Context) {
				<-ctx.Done()
				select {
				case calls <- struct{}{}:
				default:
				}
			},
			periodic.WithClock(clk),
		)
	)

	defer handle.Stop()

	clk.Add(time.Second)

	timeout := time.NewTimer(100 * time.Millisecond)
	defer timeout.Stop()

	for done := false; !done; /* noincr */ {
		select {
		case <-calls:
			require.FailNow(t, "unexpected periodic call")
		case <-timeout.C:
			done = true
		default:
			clk.Add(time.Second)
		}
	}

	cancel()

	select {
	case <-calls:
	case <-time.After(time.Second):
		require.FailNow(t, "timed out waiting for cancel call")
	}
}

func TestStartWithContext_ContextPreCanceled(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	var (
		calls  = make(chan struct{})
		clk    = clock.NewFakeClock()
		handle = periodic.StartWithContext(
			canceled,
			time.Second,
			func(ctx context.Context) {
				select {
				case <-ctx.Done():
				case calls <- struct{}{}:
				}
			},
			periodic.WithClock(clk),
		)
	)

	defer handle.Stop()

	timeout := time.NewTimer(100 * time.Millisecond)
	defer timeout.Stop()

	for seen := 0; seen < 1; /* noincr */ {
		select {
		case <-calls:
			seen++
		case <-timeout.C:
			require.Equal(t, seen, 0)
			return
		default:
			clk.Add(time.Second)
		}
	}
}

func TestHandle_Run(t *testing.T) {
	var (
		called = make(chan struct{}, 1)
		calls  atomic.Int64
		clk    = clock.NewFakeClock()
		handle = periodic.Start(
			time.Second,
			func(ctx context.Context) {
				select {
				case called <- struct{}{}:
					calls.Add(1)
				case <-ctx.Done():
				}
			},
			periodic.WithClock(clk),
		)
	)

	defer handle.Stop()

	for i := 0; i < 3; i++ {
		handle.Run()
		clk.Add(time.Second)
		requireRecvWithTimeout(t, called, time.Second)
		requireRecvWithTimeout(t, called, time.Second)
		require.EqualValues(t, (i+1)*2, calls.Load())
	}
}

func recvWithTimeout[T any](ch <-chan T, timeout time.Duration) bool {
	_, ok := channels.RecvWithTimeout(context.Background(), ch, timeout)
	return ok
}

func requireRecvWithTimeout[T any](
	t *testing.T,
	ch <-chan T,
	timeout time.Duration,
) {
	require.True(t, recvWithTimeout(ch, timeout))
}
