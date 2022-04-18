package chrono

import (
	_ "unsafe" // for go:linkname
)

// Nanotime provides the current monotonic system time as integer nanoseconds.
//go:linkname Nanotime runtime.nanotime
func Nanotime() int64
