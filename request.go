package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

type Request struct {
	client      *http.Client
	url         *url.URL
	headers     map[string]string
	queryParams url.Values
	body        interface{}
	bodyBytes   []byte
	result      interface{}
}

func New(requestURL *url.URL, httpClient *http.Client, accepts string) *Request {
	return &Request{
		client: httpClient,
		url:    requestURL,
		headers: map[string]string{
			"Accept": accepts,
		},
		queryParams: make(url.Values),
	}
}

func (r *Request) WithHeader(key string, value string) *Request {
	r.headers[key] = value
	return r
}

func (r *Request) WithQueryParam(key string, value string) *Request {
	r.queryParams.Add(key, value)
	return r
}

func (r *Request) WithJSONBody(body interface{}) *Request {
	r.headers["Content-Type"] = "application/json"
	r.body = body
	return r
}

func (r *Request) WithBodyBytes(body []byte) *Request {
	r.bodyBytes = body
	return r
}

func (r *Request) WithAuthorization(token string) *Request {
	r.headers["Authorization"] = "Bearer " + token
	return r
}

func (r *Request) WithHeaders(headers map[string]string) *Request {
	for key, value := range headers {
		r.headers[key] = value
	}
	return r
}

func (r *Request) WithQueryParams(queryParams url.Values) *Request {
	for key, value := range queryParams {
		r.queryParams[key] = value
	}
	return r
}

func (r *Request) Get(ctx context.Context, path string) (interface{}, error) {
	return r.do(ctx, http.MethodGet, path)
}

func (r *Request) Post(ctx context.Context, path string) (interface{}, error) {
	if r.body != nil {
		var err error
		r.bodyBytes, err = json.Marshal(r.body)
		if err != nil {
			return nil, err
		}
	}
	return r.do(ctx, http.MethodPost, path)
}

func (r *Request) Put(ctx context.Context, path string) (interface{}, error) {
	if r.body != nil {
		var err error
		r.bodyBytes, err = json.Marshal(r.body)
		if err != nil {
			return nil, err
		}
	}
	return r.do(ctx, http.MethodPut, path)
}

func (r *Request) Patch(ctx context.Context, path string) (interface{}, error) {
	if r.body != nil {
		var err error
		r.bodyBytes, err = json.Marshal(r.body)
		if err != nil {
			return nil, err
		}
	}
	return r.do(ctx, http.MethodPatch, path)
}

func (r *Request) Delete(ctx context.Context, path string) (interface{}, error) {
	return r.do(ctx, http.MethodDelete, path)
}

func (r *Request) WithResult(result interface{}) *Request {
	r.result = result
	return r
}

func (r *Request) Request(ctx context.Context, method, path string) (*http.Request, error) {
	ref, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	refURL := r.url.ResolveReference(ref)

	q, err := url.QueryUnescape(r.queryParams.Encode())
	if err != nil {
		return nil, err
	}
	refURL.RawQuery = q

	req, err := http.NewRequestWithContext(ctx, method, refURL.String(), bytes.NewReader(r.bodyBytes))
	if err != nil {
		return nil, err
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	if len(r.bodyBytes) > 0 {
		req.Header.Add("Content-Length", strconv.Itoa(len(r.bodyBytes)))
	}

	return req, err
}

func (r *Request) do(ctx context.Context, method, path string) (interface{}, error) {
	req, err := r.Request(ctx, method, path)
	if err != nil {
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if r.result != nil {
		err = json.NewDecoder(resp.Body).Decode(r.result)
		if err != nil {
			return nil, err
		}
	}

	return r.result, nil
}
