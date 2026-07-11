package aitaskbuilder

import "time"

// SetBatchSyncPollSleepForTesting replaces the poll sleep function for the duration of a
// test. Call the returned function (typically via defer) to restore the original.
//
// This file is compiled only during `go test` and is intentionally in
// package aitaskbuilder (not aitaskbuilder_test) so that it can access unexported
// variables while still being callable from external test packages.
func SetBatchSyncPollSleepForTesting(f func(time.Duration)) func() {
	prev := batchSyncPollSleep
	batchSyncPollSleep = f
	return func() { batchSyncPollSleep = prev }
}
