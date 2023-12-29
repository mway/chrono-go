[![GoDoc][doc-img]][doc-link] [![Build Status][ci-img]][ci-link] [![Coverage Status][cov-img]][cov-link] [![Report Card][report-img]][report-link]

[doc-img]: https://pkg.go.dev/badge/go.mway.dev/chrono
[doc-link]: https://pkg.go.dev/go.mway.dev/chrono
[ci-img]: https://github.com/mway/chrono-go/actions/workflows/go.yml/badge.svg
[ci-link]: https://github.com/mway/chrono-go/actions/workflows/go.yml
[cov-img]: https://codecov.io/gh/mway/chrono-go/branch/main/graph/badge.svg
[cov-link]: https://codecov.io/gh/mway/chrono-go
[report-img]: https://goreportcard.com/badge/go.mway.dev/chrono
[report-link]: https://goreportcard.com/report/go.mway.dev/chrono

# go.mway.dev/chrono

`chrono` is a small collection of useful time-based types and helpers. It
provides:

- **go.mway.dev/chrono**
  - An exported [`Nanotime`][nanotime-doc] function, which provides access to
    the underlying system clock (via [`runtime.nanotime`][nanotime-stdlib])
- **go.mway.dev/chrono/clock**
  - A common [`Clock`][clock-doc] interface shared by all clocks
  - Monotonic and wall `Clock`s
  - A [`ThrottledClock`][throttled-clock-doc] to provide configurable time
    memoization to reduce time-based syscalls
  - A [`FakeClock`][fake-clock-doc] implementation, to support mocking time
  - A lightweight [`Stopwatch`][stopwatch-doc] for trivially (and continuously)
    measuring elapsed time
- **go.mway.dev/periodic**
  - A [`Handle`][periodic-handler-doc] to manage functions that are run
    periodically via [`Start`][periodic-start-doc]
- **go.mway.dev/chrono/rate**
  - A [`Recorder`][recorder-doc] that performs simple, empirical rate
    calculation, optionally against a custom clock
  - A lightweight [`Rate`][rate-doc] type that provides simple translation of
    an absolute rate into different denominations

These packages are intended for general use, and as a replacement for the
long-archived [github.com/andres-erbsen/clock][erbsen-clock-repo] package.

## Getting Started

To start using a [`Clock`][clock-doc], first determine whether the goal is to
*tell* time, or to *measure* time:

- When *telling* time, wall clocks are necessary. Use [NewWallClock][new-wall-clock-doc].
- When *measuring* time, monotonic clocks are preferable, though in most cases
  a wall clock can be used as well. Use [NewMonotonicClock][new-monotonic-clock-doc].

The only difference between a wall clock and a monotonic clock is the time
source being used, which in this implementation is either a [`TimeFunc`][timefunc-doc]
or a [`NanotimeFunc`][nanotimefunc-doc].

After selecting a type of clock, simply construct and use it:

```go
package main

import (
  "time"
  
  "go.mway.dev/chrono/clock"
)

func main() {
  var (
    clk       = clock.NewWallClock()
    now       = clk.Now()       // time.Time
    nanos     = clk.Nanotime()  // int64
    stopwatch = clk.NewStopwatch()
    ticker    = clk.NewTicker(time.Second)
    timer     = clk.NewTimer(3*time.Second)
  )

  fmt.Printf("It is now: %s (~%d)\n", now, nanos)

  func() {
    for {
      select {
      case <-ticker.C:
        fmt.Println("Tick!")
      case <-timer.C:
        fmt.Println("Done!")
        return
      }
    }
  }()

  fmt.Println("Ticker/timer took", stopwatch.Elapsed())

  // etc.
}
```

The goal of [`Clock`][clock-doc] is to be comparable to using the standard
library [`time`][time-pkg-doc], i.e. any common `time`-related functions should
be provided as part of the `Clock` API.

### Examples

TODO

### Throttling

In some cases, it may be desirable to limit the number of underlying time-based
syscalls being made, for example when needing to attribute time within a tight
loop. Rather than needing to throttle or debounce such calls themselves, users
can use a [`ThrottledClock`][throttled-clock-doc], which does this at a
configurable resolution:

```go
// Use monotonic time that only updates once per second
clk := clock.NewThrottledMonotonicClock(time.Second)
defer clk.Stop() // free resources

// Issue an arbitrary number of time-based calls
for i := 0; i < 1_000_000; i++ {
  clk.Nanotime()
}
```

A background routine updates the clock's time from the configured source at
the specified interval, and time-based calls on a `ThrottledClock` use the
internally cached time.

## Contributions & Feedback

Pull requests are welcome, and all feedback is appreciated!

[time-pkg-doc]: https://pkg.go.dev/time
[time-doc]: https://pkg.go.dev/time#Time
[nanotime-doc]: https://pkg.go.dev/go.mway.dev/chrono#Nanotime
[nanotime-stdlib]: https://cs.opensource.google/go/go/+/refs/tags/go1.20.1:src/runtime/time_nofake.go;l=18-20
[clock-doc]: https://pkg.go.dev/go.mway.dev/chrono/clock#Clock
[throttled-clock-doc]: https://pkg.go.dev/go.mway.dev/chrono/clock#ThrottledClock
[periodic-handle-doc]: https://pkg.go.dev/go.mway.dev/chrono/periodic#Handle
[periodic-start-doc]: https://pkg.go.dev/go.mway.dev/chrono/periodic#Start
[fake-clock-doc]: https://pkg.go.dev/go.mway.dev/chrono/clock#FakeClock
[stopwatch-doc]: https://pkg.go.dev/go.mway.dev/chrono/clock#Stopwatch
[recorder-doc]: https://pkg.go.dev/go.mway.dev/chrono/rate#Recorder
[rate-doc]: https://pkg.go.dev/go.mway.dev/chrono/rate#Rate
[erbsen-clock-repo]: https://github.com/andres-erbsen/clock
[new-wall-clock-doc]:https://pkg.go.dev/go.mway.dev/chrono/clock#NewWallClock
[new-monotonic-clock-doc]:https://pkg.go.dev/go.mway.dev/chrono/clock#NewMonotonicClock
[timefunc-doc]:https://pkg.go.dev/go.mway.dev/chrono/clock#TimeFunc
[nanotimefunc-doc]:https://pkg.go.dev/go.mway.dev/chrono/clock#NanotimeFunc
