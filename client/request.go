package client

import (
	"encoding/json"
	"net/http"
	"net/url"
	"bytes"
	"strconv"
	"errors"
	"io"
)

type basicAuth struct {
	username string
	password string
}

type ResponseBodyHandler func(responseBody interface{}) error

type Request struct {
	client *Client
	apiCall string
	method string
	headers http.Header
	body interface{}
	responseBody interface{}
	handler ResponseBodyHandler
	parameters url.Values
	auth *basicAuth
	expectedStatusCode *int
}

func NewRequest(client *Client, method string, apiCall string) *Request {

	builder := Request{
		client: client,
		method: method,
		apiCall: apiCall,
		headers: make(http.Header),
		parameters: make(url.Values),
	}

	return &builder
}

func (r *Request) Param(name string, value string) *Request {
	r.parameters.Add(name, value)
	return r
}

func (r *Request) ParamBool(name string, value bool) *Request {
	r.parameters.Add(name, strconv.FormatBool(value))
	return r
}

func (r *Request) ParamInt(name string, value int) *Request {
	r.parameters.Add(name, strconv.Itoa(value))
	return r
}

func (r *Request) ParamInt64(name string, value int64) *Request {
	r.parameters.Add(name, strconv.FormatInt(value, 10))
	return r
}

func (r *Request) ParamUint(name string, value uint) *Request {
	return r.ParamUint64(name, uint64(value))
}

func (r *Request) ParamUint64(name string, value uint64) *Request {
	r.parameters.Add(name, strconv.FormatUint(value, 10))
	return r
}

func (r *Request) Body(content interface{}) *Request {
	r.body = content
	return r
}

func (r *Request) BasicAuth(username string, password string) *Request {
	r.auth = &basicAuth{
		username: username,
		password: password,
	}
	return r
}

func (r *Request) Expect(statusCode int) *Request {
	r.expectedStatusCode = &statusCode
	return r
}

func (r *Request) ResponseBody(content interface{}) *Request {
	r.responseBody = content
	return r
}

func (r *Request) ResponseBodyHandler(handler ResponseBodyHandler) *Request {
	r.handler = handler
	return r
}

func (r *Request) Execute() (*Response, error) {

	var reader io.Reader = nil
	
	if r.body != nil {
		body, err := json.Marshal(r.body)

		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(body)
	} 
			
	req, err := http.NewRequest(r.method,
		*r.client.url + r.apiCall, reader)

	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = r.parameters.Encode()
	req.Header = r.headers
	req.Header.Add("User-Agent", r.client.userAgent)

	if r.body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	if r.auth != nil {
		req.SetBasicAuth(r.auth.username, r.auth.password)
	} else {
		authToken, err := ReadToken()

		// If we didn't have a token, just try anyway and let
		// the API return error if we are not requesting an auth-less
		// API endpoint
		if err == nil {
			req.Header.Add("Authorization", "Bearer " + authToken.Token)
		}
	}
	
	httpRsp, err := r.client.httpClient.Do(req)

	rsp := NewResponse(httpRsp)
	
	if err != nil {
		return rsp, err
	}

	// Check if we got the status code we expected
	if r.expectedStatusCode != nil &&
		*r.expectedStatusCode != httpRsp.StatusCode {

		// Read error message if any
		errorMsg, err := rsp.ReadError()
		
		if err != nil {
			err = errors.New("Unexpected status code " + httpRsp.Status)
		} else if len(errorMsg.Errors) > 0 {
			err = errors.New("Error: " + errorMsg.Errors[0].Message)
		}
		return rsp, err
	}

	if r.responseBody != nil {
		err = rsp.Read(r.responseBody)

		if err != nil {
			return rsp, err
		}

		if r.handler != nil {
			err = r.handler(r.responseBody)
		}
	}

	return rsp, err
}
