package client

import (
	"encoding/json"
	"errors"
	"net/http"
)

type ResponseBodyReader func(*Response, interface{}) error

type Response struct {
	httpResponse *http.Response
	reader       ResponseBodyReader // Override for testing
}

type RestError struct {
	Errors []struct {
		Code    uint
		Message string
		Details string
	}
}

func NewResponse(http *http.Response) *Response {
	return &Response{
		httpResponse: http,
		reader:       defaultResponseBodyReader,
	}
}

func defaultResponseBodyReader(r *Response, into interface{}) error {
	defer r.httpResponse.Body.Close()
	return json.NewDecoder(r.httpResponse.Body).Decode(&into)
}

func (r *Response) Read(into interface{}) error {
	return r.reader(r, into)
}

func (r *Response) Http() *http.Response {
	return r.httpResponse
}

func (r *Response) ReadError() (*RestError, error) {

	errorMsg := new(RestError)

	if r.Http().ContentLength <= 0 {
		return nil, errors.New("No error in response")
	}

	err := r.Read(errorMsg)

	if err != nil {
		return nil, err
	}

	return errorMsg, nil
}
