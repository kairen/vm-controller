package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/kairen/vm-controller/pkg/client/util"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Request struct {
	*http.Request
	baseURL *url.URL
	client  HTTPClient

	token   string
	queries string
	subpath string
	method  string
	body    interface{}

	ctx context.Context
}

func NewRequest(ctx context.Context, client HTTPClient, baseURL *url.URL, method string) *Request {
	return &Request{
		method:  method,
		client:  client,
		baseURL: baseURL,
		ctx:     ctx,
	}
}

func NewRequestWithBody(ctx context.Context, client HTTPClient, baseURL *url.URL, method string, body interface{}) *Request {
	r := NewRequest(ctx, client, baseURL, method)
	r.body = body
	return r
}

func (r *Request) Context(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

func (r *Request) Body(body interface{}) *Request {
	r.body = body
	return r
}

func (r *Request) Suffix(segments ...string) *Request {
	r.subpath = path.Join(r.subpath, path.Join(segments...))
	return r
}

func (r *Request) Token(token string) *Request {
	r.token = token
	return r
}

func (r *Request) Queries(opts interface{}) *Request {
	queries, err := util.AddListOptions(r.queries, opts)
	if err == nil {
		r.queries = queries
	}
	return r
}

// FullPath returns subpath + resource + queries path
func (r *Request) FullPath() string {
	return r.subpath + r.queries
}

func (r *Request) request(fn func(*http.Request, *http.Response)) error {
	client := r.client
	if client == nil {
		client = http.DefaultClient
	}

	rel, err := url.Parse(r.FullPath())
	if err != nil {
		return err
	}

	u := r.baseURL.ResolveReference(rel)
	buf := new(bytes.Buffer)
	if r.body != nil {
		err = json.NewEncoder(buf).Encode(r.body)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(r.method, u.String(), buf)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Authorization", r.token)

	req = req.WithContext(r.ctx)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	fn(req, resp)
	return nil
}

func (r *Request) transformResponse(req *http.Request, resp *http.Response) Result {
	var body []byte
	if resp.Body != nil {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			unexpectedErr := fmt.Errorf("Unexpected error %#v when reading response body. Please retry.", err)
			return Result{err: unexpectedErr}
		}
		body = data
	}

	response := &Response{Response: resp}
	if err := json.Unmarshal(body, &response.Data); err != nil {
		return Result{err: err}
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusPartialContent {
		return Result{err: response.Error()}
	}

	return Result{resp: response}
}

func (r *Request) Do() Result {
	var result Result
	err := r.request(func(req *http.Request, resp *http.Response) {
		result = r.transformResponse(req, resp)
	})
	if err != nil {
		return Result{err: err}
	}
	return result
}

type Response struct {
	*http.Response
	Data interface{} `json:"data,omitempty"`
}

func (r *Response) Error() error {
	err := &ResponeError{
		Method:     r.Response.Request.Method,
		URL:        r.Response.Request.URL.String(),
		StatusCode: r.Response.StatusCode,
	}
	return err
}

type Result struct {
	resp *Response
	err  error
}

func (r Result) Into(v interface{}) error {
	if r.err != nil {
		return r.err
	}

	if v != nil {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(r.resp.Data); err != nil {
			return err
		}

		if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
			return err
		}
	}
	return nil
}
