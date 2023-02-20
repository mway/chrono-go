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

package chrono_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mway.dev/chrono"
)

func TestTimestamp_Add(t *testing.T) {
	ts := chrono.NewTimestampFromNanos(123)
	require.Equal(t, ts, ts.Add(0))
	require.Equal(t, ts-1, ts.Add(-1))
	require.Equal(t, ts+1, ts.Add(1))
}

func TestTimestamp_After(t *testing.T) {
	var (
		ts     = chrono.NewTimestampFromNanos(123)
		timeob = ts.Time()
	)

	require.False(t, ts.After(timeob))
	require.True(t, ts.After(timeob.Add(-1)))

	require.False(t, ts.AfterTimestamp(ts))
	require.True(t, ts.AfterTimestamp(ts-1))
}

func TestTimestamp_Before(t *testing.T) {
	var (
		ts     = chrono.NewTimestampFromNanos(123)
		timeob = ts.Time()
	)

	require.False(t, ts.Before(timeob))
	require.True(t, ts.Before(timeob.Add(1)))

	require.False(t, ts.BeforeTimestamp(ts))
	require.True(t, ts.BeforeTimestamp(ts+1))
}

func TestTimestamp_Equal(t *testing.T) {
	var (
		ts     = chrono.NewTimestampFromNanos(123)
		timeob = ts.Time()
	)

	require.True(t, ts.Equal(timeob))
	require.False(t, ts.Equal(timeob.Add(1)))
	require.False(t, ts.Equal(timeob.Add(-1)))

	require.True(t, ts.EqualTimestamp(ts))
	require.False(t, ts.EqualTimestamp(ts+1))
	require.False(t, ts.EqualTimestamp(ts-1))
}

func TestTimestamp_Format(t *testing.T) {
	ts := chrono.NewTimestampFromNanos(int64(time.Hour))
	require.Equal(t, ts.Time().Format(time.UnixDate), ts.Format(time.UnixDate))
}

func TestTimestamp_IsZero(t *testing.T) {
	require.True(t, chrono.Timestamp(0).IsZero())
	require.False(t, chrono.Timestamp(1).IsZero())
	require.False(t, chrono.Timestamp(-1).IsZero())
}

func TestTimestamp_String(t *testing.T) {
	ts := chrono.NewTimestampFromNanos(int64(time.Hour))
	require.Equal(t, ts.Time().String(), ts.String())
}

func TestTimestamp_Sub(t *testing.T) {
	var (
		ts     = chrono.NewTimestampFromNanos(123)
		timeob = ts.Time()
	)

	require.EqualValues(t, 0, ts.Sub(timeob))
	require.EqualValues(t, 1, ts.Sub(timeob.Add(-1)))
	require.EqualValues(t, -1, ts.Sub(timeob.Add(1)))

	require.EqualValues(t, 0, ts.SubTimestamp(ts))
	require.EqualValues(t, 1, ts.SubTimestamp(ts-1))
	require.EqualValues(t, -1, ts.SubTimestamp(ts+1))
}

func TestTimestamp_UnixNano(t *testing.T) {
	var (
		now    = time.Now()
		ts     = chrono.NewTimestampFromTime(now)
		timeob = ts.Time()
	)

	require.Equal(t, now.UnixNano(), ts.UnixNano())
	require.Equal(t, ts.UnixNano(), timeob.UnixNano())
}
