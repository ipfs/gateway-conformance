package test

import (
	"fmt"
	"strings"
)

func H[T any](hint string, v T) WithHint[T] {
	return WithHint[T]{v, hint}
}

func HeaderContains(hint string, v string) WithHint[Check[string]] {
	return H[Check[string]](hint, func(s string) bool {
		return strings.Contains(s, v)
	})
}

type Testable interface {
}

type HeaderIface struct {
	Key   string
	Value Testable
	Hint  string
}

type RequestBuidler struct {
	Method_  string
	Url_     string
	Headers_ map[string]string
}

func RequesT() RequestBuidler {
	return RequestBuidler{Method_: "GET"}
}

func (r RequestBuidler) Url(url string, args ...any) RequestBuidler {
	r.Url_ = fmt.Sprintf(url, args...)
	return r
}

func (r RequestBuidler) Header(h HeaderBuilder) RequestBuidler {
	if r.Headers_ == nil {
		r.Headers_ = make(map[string]string)
	}

	r.Headers_[h.Key] = h.Value.(string)
	return r
}

func (r RequestBuidler) Headers(hs ...HeaderBuilder) RequestBuidler {
	for _, h := range hs {
		r = r.Header(h)
	}

	return r
}

// generate the Request:
func (r RequestBuidler) Request() Request {
	return Request{
		Method:  r.Method_,
		Url:     r.Url_,
		Headers: r.Headers_,
	}
}

type ExpectBuilder struct {
	StatusCode int
	Headers_   []HeaderIface
	Body       []byte
}

func Expect() ExpectBuilder {
	return ExpectBuilder{}
}

func (e ExpectBuilder) Status(statusCode int) ExpectBuilder {
	e.StatusCode = statusCode
	return e
}

func (e ExpectBuilder) Header(h HeaderBuilder) ExpectBuilder {
	e.Headers_ = append(e.Headers_, h.Header())
	return e
}

func (e ExpectBuilder) Headers(hs ...HeaderBuilder) ExpectBuilder {
	xs := make([]HeaderIface, len(hs))
	for i, h := range hs {
		xs[i] = h.Header()
	}

	e.Headers_ = append(e.Headers_, xs...)
	return e
}

func (e ExpectBuilder) Response() Response {
	headers := make(map[string]interface{})

	return Response{
		StatusCode: e.StatusCode,
		Headers:    headers,
		Body:       e.Body,
	}
}

type HeaderBuilder struct {
	Key   string
	Value Testable
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
	h.Value = HeaderContains(h.Hint_, value)
	return h
}

func (h HeaderBuilder) Hint(hint string) HeaderBuilder {
	h.Hint_ = hint
	return h
}

func (h HeaderBuilder) Equals(value string, args ...any) HeaderBuilder {
	h.Value = fmt.Sprintf(value, args...)
	return h
}

func (h HeaderBuilder) IsEmpty() HeaderBuilder {
	h.Value = ""
	return h
}

func (h HeaderBuilder) Header() HeaderIface {
	return HeaderIface{
		Key:   h.Key,
		Value: h.Value,
		Hint:  h.Hint_,
	}
}
