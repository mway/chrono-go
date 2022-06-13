package stopwatch

import (
	"github.com/pkg/errors"
	"go.mway.dev/chrono/clock"
	"go.uber.org/multierr"
)

var (
	// ErrNilClock indicates that the given clock is nil.
	ErrNilClock = errors.New("nil clock provided")

	_defaultOptions = Options{
		Clock: clock.NewMonotonicClock(),
	}

	_ Option = Options{}
)

// Options configure a Stopwatch.
type Options struct {
	Clock clock.Clock
}

// DefaultOptions returns a new, default Options.
func DefaultOptions() Options {
	return _defaultOptions
}

// Validate returns an error if this Options contains invalid data.
func (o Options) Validate() (err error) {
	if o.Clock == nil {
		err = multierr.Append(err, ErrNilClock)
	}

	return
}

// With returns a new Options based on o with the given opts merged onto it.
func (o Options) With(opts ...Option) Options {
	for _, opt := range opts {
		opt.apply(&o)
	}

	return o
}

func (o Options) apply(other *Options) {
	if o.Clock != nil {
		other.Clock = o.Clock
	}
}

// An Option configures a Stopwatch.
type Option interface {
	apply(*Options)
}

// WithClock returns an Option that configures a Stopwatch to use the given
// clock.
func WithClock(clk clock.Clock) Option {
	return optionFunc(func(o *Options) {
		o.Clock = clk
	})
}

type optionFunc func(*Options)

func (f optionFunc) apply(o *Options) {
	f(o)
}
