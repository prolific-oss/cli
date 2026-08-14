package client

import (
	"fmt"
	"io"
	"net/http"
)

type ExecuteBuilder struct {
	client           *Client
	method           string
	url              string
	body             any
	response         any
	expectedStatuses []int
}

func (c *Client) ExecuteBuilder() *ExecuteBuilder {
	return &ExecuteBuilder{client: c}
}

func (b *ExecuteBuilder) Get(url string, response any) (*http.Response, error) {
	return b.GetRequest(url).Status(http.StatusOK).Decode(response).Execute()
}

// GetInto sends a GET request and decodes the response, accepting any status
// Client.Execute already treated as a success (see Execute) rather than
// asserting a specific one — for endpoints where the caller never checked
// the status code beyond that.
func (b *ExecuteBuilder) GetInto(url string, response any) (*http.Response, error) {
	return b.GetRequest(url).Decode(response).Execute()
}

func (b *ExecuteBuilder) GetRequest(url string) *ExecuteBuilder {
	return b.newRequest(http.MethodGet, url)
}

func (b *ExecuteBuilder) PostRequest(url string) *ExecuteBuilder {
	return b.newRequest(http.MethodPost, url)
}

func (b *ExecuteBuilder) PatchRequest(url string) *ExecuteBuilder {
	return b.newRequest(http.MethodPatch, url)
}

func (b *ExecuteBuilder) PutRequest(url string) *ExecuteBuilder {
	return b.newRequest(http.MethodPut, url)
}

func (b *ExecuteBuilder) DeleteRequest(url string) *ExecuteBuilder {
	return b.newRequest(http.MethodDelete, url)
}

// newRequest clears any state carried over from a previous request, so a
// builder that is held and reused cannot send an earlier request's body or
// accept an earlier request's statuses.
func (b *ExecuteBuilder) newRequest(method, url string) *ExecuteBuilder {
	b.method = method
	b.url = url
	b.body = nil
	b.response = nil
	b.expectedStatuses = nil
	return b
}

func (b *ExecuteBuilder) Body(v any) *ExecuteBuilder {
	b.body = v
	return b
}

func (b *ExecuteBuilder) Decode(v any) *ExecuteBuilder {
	b.response = v
	return b
}

func (b *ExecuteBuilder) Status(statuses ...int) *ExecuteBuilder {
	b.expectedStatuses = statuses
	return b
}

func (b *ExecuteBuilder) Execute() (*http.Response, error) {
	httpResponse, err := b.client.Execute(b.method, b.url, b.body, b.response)

	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", b.url, err)
	}

	// Status was never called: Client.Execute already treated anything below
	// 400 as a success, so there's nothing further to assert here.
	if len(b.expectedStatuses) == 0 {
		return httpResponse, nil
	}

	// Check if the status code is one of the accepted ones
	for _, status := range b.expectedStatuses {
		if httpResponse.StatusCode == status {
			return httpResponse, nil
		}
	}

	// Status didn't match — read body for error message
	body, _ := io.ReadAll(httpResponse.Body)
	return nil, fmt.Errorf("unexpected status code %d: %s", httpResponse.StatusCode, string(body))
}
