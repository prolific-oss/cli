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
	return b.GetRequest(url).Status(200).Decode(response).Execute()
}

func (b *ExecuteBuilder) GetRequest(url string) *ExecuteBuilder {
	b.method = http.MethodGet
	b.url = url
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
