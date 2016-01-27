package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/iobeam/iobeam/config"
)

// see time.Parse docs for why this is the case
const (
	tokenTimeForm    = "2006-01-02 15:04:05 -0700"
	contentTypeJson  = "application/json"
	contentTypePlain = "text/plain"
)

type basicAuth struct {
	username string
	password string
}

// ResponseBodyHandler is a function called when an API request returns.
// The function takes an interface{} (usually converted to JSON) and returns
// an error.
type ResponseBodyHandler func(responseBody interface{}) error

// Request is an API HTTP request to the iobeam backend.
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

// NewRequest creates a new API request to iobeam. It takes a *Client that
// handles the actual execution, a string for the HTTP method, and a string
// for the API path.
func NewRequest(client *Client, method, apiCall string) *Request {

	builder := Request{
		client:     client,
		method:     method,
		apiCall:    apiCall,
		headers:    make(http.Header),
		parameters: make(url.Values),
	}

	return &builder
}

// Param adds a query parameter to the *Request whose value is a string.
// It returns the *Request so it can be chained.
func (r *Request) Param(name, value string) *Request {
	r.parameters.Add(name, value)
	return r
}

// ParamBool adds a query parameter to the *Request whose value is a bool.
// It returns the *Request so it can be chained.
func (r *Request) ParamBool(name string, value bool) *Request {
	r.parameters.Add(name, strconv.FormatBool(value))
	return r
}

// ParamInt adds a query parameter to the *Request whose value is an int.
// It returns the *Request so it can be chained.
func (r *Request) ParamInt(name string, value int) *Request {
	r.parameters.Add(name, strconv.Itoa(value))
	return r
}

// ParamInt64 adds a query parameter to the *Request whose value is an int64.
// It returns the *Request so it can be chained.
func (r *Request) ParamInt64(name string, value int64) *Request {
	r.parameters.Add(name, strconv.FormatInt(value, 10))
	return r
}

// ParamUint adds a query parameter to the *Request whose value is a uint.
// It returns the *Request so it can be chained.
func (r *Request) ParamUint(name string, value uint) *Request {
	return r.ParamUint64(name, uint64(value))
}

// ParamUint64 adds a query parameter to the *Request whose value is a uint64.
// It returns the *Request so it can be chained.
func (r *Request) ParamUint64(name string, value uint64) *Request {
	r.parameters.Add(name, strconv.FormatUint(value, 10))
	return r
}

// Body sets the HTTP body of the *Request. It assumes the content can be
// serialized to JSON. It returns the *Request so it can be chained.
func (r *Request) Body(content interface{}) *Request {
	r.headers.Add("Content-Type", contentTypeJson)
	r.body = content
	return r
}

// BasicAuth sets the username and password to use in case of Basic authentication
// on the API endpoint. It returns the *Request so it can be chained.
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

// UserToken returns a request with user token set. It returns the *Request so it can be chained.
func (r *Request) UserToken(p *config.Profile) *Request {
	r.token, _ = ReadUserToken(p)
	return r
}

// ProjectToken returns a Request with the project token set. If the token is expired,
// an attempt to refresh the token is made. It returns the *Request so it can be chained.
func (r *Request) ProjectToken(p *config.Profile, id uint64) *Request {
	r.token, _ = ReadProjToken(p, id)
	if r.token == nil {
		return r
	}
	expired, err := r.token.IsExpired()
	if err != nil {
		return r
	}

	if expired {
		r.token, _ = r.token.Refresh(r.client, p)
	}
	return r
}

// Expect sets the HTTP status code value that the executed r should expect.
// It returns the *Request so it can be chained.
func (r *Request) Expect(statusCode int) *Request {
	r.expectedStatusCode = &statusCode
	return r
}

// ResponseBody sets the object where the *Request's response should be stored.
// It returns the *Request so it can be chained.
func (r *Request) ResponseBody(content interface{}) *Request {
	r.responseBody = content
	return r
}

// ResponseBodyHandler sets the ResponseBodyHandler to use when the request returns.
// It returns the *Request so it can be chained.
func (r *Request) ResponseBodyHandler(handler ResponseBodyHandler) *Request {
	r.handler = handler
	return r
}

// Execute causes the API request to be carried out, returning a *Response and
// possibly an error.
func (r *Request) Execute() (*Response, error) {

	var reader io.Reader

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

	// If basic auth nor a token is set, we'll try anyway and then fail as unauthorized.
	if r.auth != nil {
		req.SetBasicAuth(r.auth.username, r.auth.password)
	} else if r.token != nil {
		req.Header.Add("Authorization", "Bearer "+r.token.Token)
	}

	httpRsp, err := r.client.httpClient.Do(req)

	rsp := NewResponse(httpRsp)
	contentType := contentTypeJson
	if len(httpRsp.Header["Content-Type"]) > 0 {
		contentType = httpRsp.Header["Content-Type"][0]
	}
	contentTypeArgs := strings.Split(contentType, ";")

	caseInsensitiveEqual := func(given, want string) bool {
		return strings.ToLower(strings.TrimSpace(given)) == want
	}
	isJson := false
	isPlain := false
	if len(contentTypeArgs) == 1 {
		isJson = caseInsensitiveEqual(contentTypeArgs[0], contentTypeJson)
		isPlain = caseInsensitiveEqual(contentTypeArgs[0], contentTypePlain)
	} else if len(contentTypeArgs) == 2 {
		isUtf8 := caseInsensitiveEqual(contentTypeArgs[1], "charset=utf-8")
		if isUtf8 {
			isJson = caseInsensitiveEqual(contentTypeArgs[0], contentTypeJson)
			isPlain = caseInsensitiveEqual(contentTypeArgs[0], contentTypePlain)
		}
	}

	if !isJson && !isPlain {
		return rsp, errors.New("Unknown content-type: " + contentType)
	}

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

	var res interface{}
	if isJson && r.responseBody != nil {
		err = rsp.ReadJson(r.responseBody)
		res = r.responseBody
	} else if !isJson && isPlain {
		var temp string
		err = rsp.ReadPlain(&temp)
		res = temp
	} else if !isJson && !isPlain {
		return rsp, errors.New("Unknown content-type, cannot read response: " + contentType)
	}

	if err != nil {
		return rsp, err
	}

	if r.handler != nil {
		err = r.handler(res)
	}

	return rsp, err
}
