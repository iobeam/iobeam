package client

import (
	"net/http"
)

type Client struct {
	httpClient  *http.Client 
	url *string
	username string
	password string
	authToken string
	userAgent string
}

func NewClient(target *string, username string, password string) (*Client) {

	client := Client {
		httpClient: &http.Client {},
		url: target,
		username: username,
		password: password,
		userAgent: "Beam 0.1",
	}

	return &client
}

func (client *Client) Get(apiCall string) (*Request) {
	return NewRequest(client, "GET", apiCall)
}

func (client *Client) Post(apiCall string) (*Request) {
	return NewRequest(client, "POST", apiCall)
}

func (client *Client) Put(apiCall string) (*Request) {
	return NewRequest(client, "PUT", apiCall)
}

func (client *Client) Patch(apiCall string) (*Request) {
	return NewRequest(client, "PATCH", apiCall)
}

func (client *Client) Delete(apiCall string) (*Request) {
	return NewRequest(client, "DELETE", apiCall)
}
