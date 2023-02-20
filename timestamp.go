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

package chrono

import (
	"time"
)

// A Timestamp is an integer timestamp.
//
// Timestamp contains various methods that accept or return a time.Time, which
// encodes timezone information; users must ensure that any time values and
// timezones are appropriate to use for those functions: a Timestamp and
// time.Time are expected to be in the same timezone.
type Timestamp int64

// NewTimestampFromTime returns x as a [Timestamp].
func NewTimestampFromTime(x time.Time) Timestamp {
	return NewTimestampFromNanos(x.UnixNano())
}

// NewTimestampFromNanos returns x as a [Timestamp].
func NewTimestampFromNanos(x int64) Timestamp {
	return Timestamp(x)
}

// Add adds d to t and returns the resulting value.
func (t Timestamp) Add(d time.Duration) Timestamp {
	return t + Timestamp(d)
}

// After returns whether u is after t.
func (t Timestamp) After(u time.Time) bool {
	return int64(t) > u.UnixNano()
}

// AfterTimestamp returns whether u is after t.
func (t Timestamp) AfterTimestamp(u Timestamp) bool {
	return t > u
}

// Before returns whether u is before t.
func (t Timestamp) Before(u time.Time) bool {
	return int64(t) < u.UnixNano()
}

// BeforeTimestamp returns whether u is before t.
func (t Timestamp) BeforeTimestamp(u Timestamp) bool {
	return t < u
}

// Equal returns whether u refers to the same time as t.
func (t Timestamp) Equal(u time.Time) bool {
	return int64(t) == u.UnixNano()
}

// EqualTimestamp returns whether u refers to the same time as t.
func (t Timestamp) EqualTimestamp(u Timestamp) bool {
	return t == u
}

// Format formats t based on the given layout.
func (t Timestamp) Format(layout string) string {
	return t.Time().Format(layout)
}

// IsZero returns whether t is a zero value.
func (t Timestamp) IsZero() bool {
	return t == 0
}

// String returns t as a timestamp string.
func (t Timestamp) String() string {
	return t.Time().String()
}

// Sub subtracts u from t and returns the resulting value.
func (t Timestamp) Sub(u time.Time) time.Duration {
	return time.Duration(int64(t) - u.UnixNano())
}

// SubTimestamp subtracts u from t and returns the resulting value.
func (t Timestamp) SubTimestamp(u Timestamp) time.Duration {
	return time.Duration(t - u)
}

// Time converts t to a UTC-based [time.Time].
func (t Timestamp) Time() time.Time {
	return time.Unix(0, int64(t))
}

// UnixNano returns t as integer nanoseconds.
func (t Timestamp) UnixNano() int64 {
	return int64(t)
}
