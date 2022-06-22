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

import (
	"time"
)

// A FakeHook is used to inject callbacks on a FakeClock when timers, tickers,
// or both are created, reset, or stopped.
type FakeHook struct {
	// Filter controls which type(s) are affected by this hook.
	Filter FakeHookFilter
	// OnCreate will be called when type(s) matched by Filter are created.
	OnCreate func(*FakeClock, time.Duration)
	// OnCreate will be called when type(s) matched by Filter are reset.
	OnReset func(*FakeClock, time.Duration)
	// OnCreate will be called when type(s) matched by Filter are stopped.
	OnStop func(*FakeClock)
}

type resolvedHooks struct {
	timerCreate  []func(*FakeClock, time.Duration)
	timerReset   []func(*FakeClock, time.Duration)
	timerStop    []func(*FakeClock)
	tickerCreate []func(*FakeClock, time.Duration)
	tickerReset  []func(*FakeClock, time.Duration)
	tickerStop   []func(*FakeClock)
}

func (h *resolvedHooks) OnTimerCreate(clk *FakeClock, dur time.Duration) {
	for _, fn := range h.timerCreate {
		fn(clk, dur)
	}
}

func (h *resolvedHooks) OnTimerReset(clk *FakeClock, dur time.Duration) {
	for _, fn := range h.timerReset {
		fn(clk, dur)
	}
}

func (h *resolvedHooks) OnTimerStop(clk *FakeClock) {
	for _, fn := range h.timerStop {
		fn(clk)
	}
}

func (h *resolvedHooks) OnTickerCreate(clk *FakeClock, dur time.Duration) {
	for _, fn := range h.tickerCreate {
		fn(clk, dur)
	}
}

func (h *resolvedHooks) OnTickerReset(clk *FakeClock, dur time.Duration) {
	for _, fn := range h.tickerReset {
		fn(clk, dur)
	}
}

func (h *resolvedHooks) OnTickerStop(clk *FakeClock) {
	for _, fn := range h.tickerStop {
		fn(clk)
	}
}

func (h *resolvedHooks) addOnCreate(
	filter FakeHookFilter,
	callback func(*FakeClock, time.Duration),
) {
	switch filter {
	case FilterTickers:
		h.tickerCreate = append(h.tickerCreate, callback)
	case FilterTimers:
		h.timerCreate = append(h.timerCreate, callback)
	case FilterAll:
		h.tickerCreate = append(h.tickerCreate, callback)
		h.timerCreate = append(h.timerCreate, callback)
	default:
		// nop
	}
}

func (h *resolvedHooks) addOnReset(
	filter FakeHookFilter,
	callback func(*FakeClock, time.Duration),
) {
	switch filter {
	case FilterTickers:
		h.tickerReset = append(h.tickerReset, callback)
	case FilterTimers:
		h.timerReset = append(h.timerReset, callback)
	case FilterAll:
		h.tickerReset = append(h.tickerReset, callback)
		h.timerReset = append(h.timerReset, callback)
	default:
		// nop
	}
}

func (h *resolvedHooks) addOnStop(
	filter FakeHookFilter,
	callback func(*FakeClock),
) {
	switch filter {
	case FilterTickers:
		h.tickerStop = append(h.tickerStop, callback)
	case FilterTimers:
		h.timerStop = append(h.timerStop, callback)
	case FilterAll:
		h.tickerStop = append(h.tickerStop, callback)
		h.timerStop = append(h.timerStop, callback)
	default:
		// nop
	}
}

func newResolvedHooks(hooks []FakeHook) resolvedHooks {
	var resolved resolvedHooks

	for _, hook := range hooks {
		if hook.OnCreate != nil {
			resolved.addOnCreate(hook.Filter, hook.OnCreate)
		}

		if hook.OnReset != nil {
			resolved.addOnReset(hook.Filter, hook.OnReset)
		}

		if hook.OnStop != nil {
			resolved.addOnStop(hook.Filter, hook.OnStop)
		}
	}

	return resolved
}
