package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iobeam/iobeam/config"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// see time.Parse docs for why this is the case
const tokenTimeForm = "2006-01-02 15:04:05 -0700"

type basicAuth struct {
	username string
	password string
}

type ResponseBodyHandler func(responseBody interface{}) error

type Request struct {
	client             *Client
	apiCall            string
	method             string
	headers            http.Header
	body               interface{}
	responseBody       interface{}
	handler            ResponseBodyHandler
	parameters         url.Values
	auth               *basicAuth
	token              *AuthToken
	expectedStatusCode *int
}

func NewRequest(client *Client, method string, apiCall string) *Request {

	builder := Request{
		client:     client,
		method:     method,
		apiCall:    apiCall,
		headers:    make(http.Header),
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

// refreshToken takes care of refreshing the project token when it expires.
func refreshToken(r *Request, t *AuthToken, p *config.Profile) {
	type data struct {
		OldToken string `json:"refresh_token"`
	}
	body := data{OldToken: t.Token}
	client := r.client
	reqPath := "/v1/tokens/project"
	_, _ = client.Post(reqPath).
		Expect(200).
		Body(body).
		ResponseBody(new(AuthToken)).
		ResponseBodyHandler(func(token interface{}) error {

		projToken := token.(*AuthToken)
		err := projToken.Save(p)
		if err != nil {
			fmt.Printf("Could not save new token: %s\n", err)
		}

		err = p.UpdateActiveProject(projToken.ProjectId)
		if err != nil {
			fmt.Printf("Could not update active project: %s\n", err)
		}
		fmt.Println("New project token acquired...")
		fmt.Printf("Expires: %v\n", projToken.Expires)
		fmt.Println("-----")

		return err
	}).Execute()
}

func (r *Request) UserToken(p *config.Profile) *Request {
	r.token, _ = ReadUserToken(p)
	return r
}

// ProjectToken returns a Request with the project token set. If the token is expired,
// an attempt to refresh the token is made. This function can be chained.
func (r *Request) ProjectToken(p *config.Profile, id uint64) *Request {
	r.token, _ = ReadProjToken(p, id)
	if r.token == nil {
		return r
	}

	exp, err := time.Parse(tokenTimeForm, r.token.Expires)
	if err != nil {
		return r
	}

	now := time.Now()
	if now.After(exp) {
		refreshToken(r, r.token, p)
		r.token, _ = ReadProjToken(p, id)
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
		*r.client.url+r.apiCall, reader)

	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = r.parameters.Encode()
	req.Header = r.headers
	req.Header.Add("User-Agent", r.client.userAgent)

	if r.body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	// If basic auth nor a token is set, we'll try anyway and then fail as unauthorized.
	if r.auth != nil {
		req.SetBasicAuth(r.auth.username, r.auth.password)
	} else if r.token != nil {
		req.Header.Add("Authorization", "Bearer "+r.token.Token)
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
