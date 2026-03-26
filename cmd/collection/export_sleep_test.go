package collection

import "time"

// SetPollSleepForTesting replaces the poll sleep function for the duration of a
// test. Call the returned function (typically via defer) to restore the original.
//
// This file is compiled only during `go test` and is intentionally in
// package collection (not collection_test) so that it can access the unexported
// pollSleep variable while still being callable from external test packages.
func SetPollSleepForTesting(f func(time.Duration)) func() {
	prev := pollSleep
	pollSleep = f
	return func() { pollSleep = prev }
}
