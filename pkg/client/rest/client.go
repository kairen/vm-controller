package rest

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

const (
	version   = "0.1.0"
	userAgent = "vm/" + version
	mediaType = "application/json"
)

type Interface interface {
	Get() *Request
	Post(body interface{}) *Request
	Put(body interface{}) *Request
	Delete(body interface{}) *Request
}

// RESTClient manages communication with Kao API.
type RESTClient struct {
	baseURL *url.URL
	ctx     context.Context
	Client  *http.Client
}

func NewRESTClient(config *Config, httpClient *http.Client) (*RESTClient, error) {
	url, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(url.Path, "/") {
		url.Path += "/"
	}

	url.RawQuery = ""
	url.Fragment = ""
	return &RESTClient{
		Client:  httpClient,
		baseURL: url,
		ctx:     context.TODO(),
	}, nil
}

func (c *RESTClient) newRequest(method string, body interface{}) *Request {
	var req *Request
	switch method {
	case http.MethodGet:
		req = NewRequest(c.ctx, c.Client, c.baseURL, method)
	case http.MethodPost, http.MethodPut, http.MethodDelete:
		req = NewRequestWithBody(c.ctx, c.Client, c.baseURL, method, body)
	}
	return req
}

func (c *RESTClient) Get() *Request {
	return c.newRequest(http.MethodGet, nil)
}

func (c *RESTClient) Post(body interface{}) *Request {
	return c.newRequest(http.MethodPost, body)
}

func (c *RESTClient) Put(body interface{}) *Request {
	return c.newRequest(http.MethodPut, body)
}

func (c *RESTClient) Delete(body interface{}) *Request {
	return c.newRequest(http.MethodDelete, body)
}
