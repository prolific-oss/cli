package aitaskbuilder

import (
	"net/http"
	"time"
)

// SetBatchExportPollSleepForTesting replaces the poll sleep function for the duration of a
// test. Call the returned function (typically via defer) to restore the original.
//
// This file is compiled only during `go test` and is intentionally in
// package aitaskbuilder (not aitaskbuilder_test) so that it can access unexported
// variables while still being callable from external test packages.
func SetBatchExportPollSleepForTesting(f func(time.Duration)) func() {
	prev := batchExportPollSleep
	batchExportPollSleep = f
	return func() { batchExportPollSleep = prev }
}

// SetBatchExportDownloadClientForTesting replaces the HTTP client used for ZIP downloads.
// Use httptest.NewTLSServer and pass srv.Client() to supply a client that trusts
// the test server's self-signed certificate. Call the returned function to restore.
func SetBatchExportDownloadClientForTesting(c *http.Client) func() {
	prev := batchExportDownloadClient
	batchExportDownloadClient = c
	return func() { batchExportDownloadClient = prev }
}
