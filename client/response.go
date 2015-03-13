package client

import (
	"net/http"
	"encoding/json"
	"errors"
//	"io/ioutil"
//	"fmt"
)

type Response struct {
	httpResponse *http.Response
}

type RestError struct {
	Errors []struct {
		Code      uint
		Message   string
		Details   string
	}
}

func (r *Response) Read(into interface{}) error {
	defer r.httpResponse.Body.Close()
	//content, _ := ioutil.ReadAll(r.httpResponse.Body)
	//fmt.Printf("body=%s\n", string(content))
	return json.NewDecoder(r.httpResponse.Body).Decode(&into)
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
