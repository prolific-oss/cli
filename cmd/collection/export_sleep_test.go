package collection

import (
	"net/http"
	"time"
)

// SetPollSleepForTesting replaces the poll sleep function for the duration of a
// test. Call the returned function (typically via defer) to restore the original.
//
// This file is compiled only during `go test` and is intentionally in
// package collection (not collection_test) so that it can access unexported
// variables while still being callable from external test packages.
func SetPollSleepForTesting(f func(time.Duration)) func() {
	prev := pollSleep
	pollSleep = f
	return func() { pollSleep = prev }
}

// SetDownloadClientForTesting replaces the HTTP client used for ZIP downloads.
// Use httptest.NewTLSServer and pass srv.Client() to supply a client that trusts
// the test server's self-signed certificate. Call the returned function to restore.
func SetDownloadClientForTesting(c *http.Client) func() {
	prev := downloadClient
	downloadClient = c
	return func() { downloadClient = prev }
}
