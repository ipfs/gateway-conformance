package test

import (
	"fmt"
	"net/url"

	"github.com/ipfs/gateway-conformance/tooling/check"
)

type RequestBuilder struct {
	Method_  string
	Url_     string
	Headers_ map[string]string
	Query_   url.Values
}

func Request() RequestBuilder {
	return RequestBuilder{Method_: "GET"}
}

func (r RequestBuilder) Url(url string, args ...any) RequestBuilder {
	r.Url_ = fmt.Sprintf(url, args...)
	return r
}

func (r RequestBuilder) Query(key, value string) RequestBuilder {
	r.Query_.Add(key, value)
	return r
}

func (r RequestBuilder) Header(h HeaderBuilder) RequestBuilder {
	if r.Headers_ == nil {
		r.Headers_ = make(map[string]string)
	}

	r.Headers_[h.Key] = h.Value
	return r
}

func (r RequestBuilder) Headers(hs ...HeaderBuilder) RequestBuilder {
	for _, h := range hs {
		r = r.Header(h)
	}

	return r
}

func (r RequestBuilder) Request() CRequest {
	url := r.Url_
	if r.Query_.Encode() != "" {
		url += "?" + r.Query_.Encode()
	}

	return CRequest{
		Method:  r.Method_,
		Url:     url,
		Headers: r.Headers_,
	}
}

type ExpectBuilder struct {
	StatusCode int
	Headers_   []HeaderBuilder
	Body_      interface{}
}

func Expect() ExpectBuilder {
	return ExpectBuilder{}
}

func (e ExpectBuilder) Status(statusCode int) ExpectBuilder {
	e.StatusCode = statusCode
	return e
}

func (e ExpectBuilder) Header(h HeaderBuilder) ExpectBuilder {
	e.Headers_ = append(e.Headers_, h)
	return e
}

func (e ExpectBuilder) Body(body []byte) ExpectBuilder {
	e.Body_ = body
	return e
}

func (e ExpectBuilder) Bytes(body string) ExpectBuilder {
	e.Body_ = []byte(body)
	return e
}

func (e ExpectBuilder) Headers(hs ...HeaderBuilder) ExpectBuilder {
	e.Headers_ = append(e.Headers_, hs...)
	return e
}

func (e ExpectBuilder) Response() CResponse {
	headers := make(map[string]interface{})

	// TODO: detect collision in keys
	for _, h := range e.Headers_ {
		if h.Hint_ != "" {
			headers[h.Key] = check.WithHint(h.Hint_, h.Check)
		} else {
			headers[h.Key] = h.Check
		}
	}

	return CResponse{
		StatusCode: e.StatusCode,
		Headers:    headers,
		Body:       e.Body_,
	}
}

type HeaderBuilder struct {
	Key   string
	Value string
	Check check.Check[string]
	Hint_ string
}

func Header(key string, opts ...string) HeaderBuilder {
	if len(opts) > 1 {
		panic("too many options")
	}
	if len(opts) > 0 {
		return HeaderBuilder{Key: key, Value: opts[0], Check: check.IsEqual(opts[0])}
	}

	return HeaderBuilder{Key: key}
}

func (h HeaderBuilder) Contains(value string) HeaderBuilder {
	h.Check = check.Contains(value)
	return h
}

func (h HeaderBuilder) Matches(value string, rest ...any) HeaderBuilder {
	h.Check = check.Matches(value, rest...)
	return h
}

func (h HeaderBuilder) Hint(hint string) HeaderBuilder {
	h.Hint_ = hint
	return h
}

func (h HeaderBuilder) Equals(value string, args ...any) HeaderBuilder {
	h.Check = check.IsEqual(value, args...)
	return h
}

func (h HeaderBuilder) IsEmpty() HeaderBuilder {
	h.Check = check.CheckIsEmpty{}
	return h
}

func (h HeaderBuilder) Checks(f func(string) bool) HeaderBuilder {
	h.Check = check.CheckFunc[string]{
		Fn: f,
	}
	return h
}
