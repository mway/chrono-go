package stopwatch

import (
	"time"

	"go.mway.dev/chrono/clock"
	"go.uber.org/atomic"
)

// A Stopwatch measures elapsed time.
type Stopwatch struct {
	clock clock.Clock
	start atomic.Int64
}

// New creates a new Stopwatch with the given options.
func New(opts ...Option) (*Stopwatch, error) {
	options := DefaultOptions().With(opts...)
	if err := options.Validate(); err != nil {
		return nil, err
	}

	s := &Stopwatch{
		clock: options.Clock,
	}
	s.Reset()

	return s, nil
}

// Elapsed returns the time elapsed since the last call to Reset.
func (s *Stopwatch) Elapsed() time.Duration {
	return time.Duration(s.clock.Nanotime() - s.start.Load())
}

// Reset sets the Stopwatch's internal time to the current time.
func (s *Stopwatch) Reset() {
	s.start.Store(s.clock.Nanotime())
}
