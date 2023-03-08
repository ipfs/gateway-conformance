package test

import (
	"fmt"

	"github.com/ipfs/gateway-conformance/tooling/check"
)

type RequestBuilder struct {
	Method_    string
	Path_      string
	Subdomain_ string
	Headers_   map[string]string
}

func Request() RequestBuilder {
	return RequestBuilder{Method_: "GET"}
}

func (r RequestBuilder) Subdomain(base string, args ...any) RequestBuilder {
	r.Subdomain_ = fmt.Sprintf(base, args...)
	return r
}

func (r RequestBuilder) Path(url string, args ...any) RequestBuilder {
	r.Path_ = fmt.Sprintf(url, args...)
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
	var url = ""
	if r.Subdomain_ != "" {
		url = fmt.Sprintf("%s://%s.%s", SubdomainGatewayScheme, r.Subdomain_, SubdomainGatewayHost)

		if r.Path_ != "" {
			url = fmt.Sprintf("%s/%s", url, r.Path_)
		}
	}

	var path = ""
	if url == "" {
		path = r.Path_
	}

	return CRequest{
		Method:  r.Method_,
		Path:    path,
		Url:     url,
		Headers: r.Headers_,
	}
}

type ExpectBuilder struct {
	StatusCode int
	Headers_   []HeaderBuilder
	Body_      []byte
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

func (e ExpectBuilder) Headers(hs ...HeaderBuilder) ExpectBuilder {
	e.Headers_ = append(e.Headers_, hs...)
	return e
}

func (e ExpectBuilder) Body(body interface{}) ExpectBuilder {
	switch body := body.(type) {
	case string:
		e.Body_ = []byte(body)
	case []byte:
		e.Body_ = body
	default:
		panic("body must be string or []byte")
	}

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
		return HeaderBuilder{Key: key, Value: opts[0]}
	}

	return HeaderBuilder{Key: key}
}

func (h HeaderBuilder) Contains(value string) HeaderBuilder {
	h.Check = check.Contains(value)
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
