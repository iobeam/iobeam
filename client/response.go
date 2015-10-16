package client

import (
	"encoding/json"
	"errors"
	"io/ioutil"
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
	d := json.NewDecoder(r.httpResponse.Body)
	d.UseNumber()
	return d.Decode(&into)
}

func plainResponseBodyReader(r *Response, into *string) error {
	defer r.httpResponse.Body.Close()
	contents, err := ioutil.ReadAll(r.httpResponse.Body)
	if err != nil {
		return err
	}
	*into = string(contents)

	return nil
}

func (r *Response) Read(into interface{}) error {
	return r.reader(r, into)
}

func (r *Response) ReadJson(into interface{}) error {
	return defaultResponseBodyReader(r, into)
}

func (r *Response) ReadPlain(result *string) error {
	return plainResponseBodyReader(r, result)
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
