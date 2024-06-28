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

// Package periodic provides helpers for doing periodic work.
package periodic

import (
	"context"
	"sync"
	"time"

	"go.mway.dev/chrono/clock"
)

// A Func is a function that can be run periodically. A Func must abide by ctx.
type Func = func(ctx context.Context)

// A Handle manages a [Func] that is running periodically.
type Handle struct {
	fn     Func
	ctx    context.Context
	cancel context.CancelFunc
	clock  clock.Clock
	wg     sync.WaitGroup
}

// Start applies the given options and starts running fn every period until
// [Handle.Stop] is called. If the given period is <=0, fn will be executed
// repeatedly without any delay.
func Start(period time.Duration, fn Func, opts ...StartOption) *Handle {
	return StartWithContext(context.Background(), period, fn, opts...)
}

// StartWithContext applies the given options and starts running fn every
// period until ctx expires or [Handle.Stop] is called. If the given period is
// <=0, fn will be executed repeatedly without any delay.
func StartWithContext(
	ctx context.Context,
	period time.Duration,
	fn Func,
	opts ...StartOption,
) *Handle {
	var (
		options      = defaultStartOptions().With(opts...)
		hctx, cancel = context.WithCancel(ctx)
		h            = &Handle{
			fn:     fn,
			ctx:    hctx,
			cancel: cancel,
			clock:  options.Clock,
		}
		ready = make(chan struct{})
	)

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.runLoop(period, ready)
	}()

	<-ready
	return h
}

// Run runs the underlying [Func] with h's own [context.Context]. This call
// does not affect the period at which h is already calling the func.
func (h *Handle) Run() {
	h.RunWithContext(h.ctx)
}

// RunWithContext runs the underlying [Func] with ctx. This call does not
// affect the period at which h is already calling the func.
func (h *Handle) RunWithContext(ctx context.Context) {
	h.fn(ctx)
}

// Stop stops the [Func] being managed by h and waits for it to exit.
func (h *Handle) Stop() {
	h.cancel()
	h.wg.Wait()
}

func (h *Handle) runLoop(period time.Duration, ready chan<- struct{}) {
	var tick <-chan time.Time
	if period > 0 {
		ticker := h.clock.NewTicker(period)
		defer ticker.Stop()
		tick = ticker.C
	} else {
		tmp := make(chan time.Time)
		close(tmp)
		tick = tmp
	}

	close(ready)

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-tick:
			select {
			case <-h.ctx.Done():
				return
			default:
			}

			h.RunWithContext(h.ctx)
		}
	}
}
