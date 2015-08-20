package client

import (
	"net/http"
	"runtime"
)

const defaultUserAgent = "iobeam cli "

// Client represents a http.Client that will be used to talk to a server
// running the iobeam API.
type Client struct {
	httpClient *http.Client
	url        *string
	userAgent  string
}

// NewClient returns a new HTTP client capable of communicating with a server
// running the iobeam API. The server address is passed via target.
func NewClient(target *string, clientVersion string) *Client {

	client := Client{
		httpClient: &http.Client{},
		url:        target,
		userAgent:  defaultUserAgent + clientVersion + "-" + runtime.GOOS + "-" + runtime.GOARCH,
	}

	return &client
}

// Get returns a new Request with HTTP method of GET for the supplied resource.
func (client *Client) Get(apiCall string) *Request {
	return NewRequest(client, "GET", apiCall)
}

// Post returns a new Request with HTTP method of POST for the supplied resource.
func (client *Client) Post(apiCall string) *Request {
	return NewRequest(client, "POST", apiCall)
}

// Put returns a new Request with HTTP method of Put for the supplied resource.
func (client *Client) Put(apiCall string) *Request {
	return NewRequest(client, "PUT", apiCall)
}

// Patch returns a new Request with HTTP method of Patch for the supplied resource.
func (client *Client) Patch(apiCall string) *Request {
	return NewRequest(client, "PATCH", apiCall)
}

// Delete returns a new Request with HTTP method of Delete for the supplied resource.
func (client *Client) Delete(apiCall string) *Request {
	return NewRequest(client, "DELETE", apiCall)
}
