package test

import (
	"fmt"
	"net/url"

	"github.com/ipfs/gateway-conformance/tooling/check"
)

type RequestBuilder struct {
	Method_               string          `json:"method,omitempty"`
	Path_                 string          `json:"path,omitempty"`
	URL_                  string          `json:"url,omitempty"`
	Proxy_                string          `json:"proxy,omitempty"`
	UseProxyTunnel_       bool            `json:"useProxyTunnel,omitempty"`
	Headers_              []HeaderBuilder `json:"headers,omitempty"`
	DoNotFollowRedirects_ bool            `json:"doNotFollowRedirects,omitempty"`
	Query_                url.Values      `json:"query,omitempty"`
	Body_                 []byte          `json:"body,omitempty"`
}

func Request() RequestBuilder {
	return RequestBuilder{Method_: "GET",
		Query_: make(url.Values)}
}

func (r RequestBuilder) Path(path string, args ...any) RequestBuilder {
	r.Path_ = fmt.Sprintf(path, args...)
	return r
}

func (r RequestBuilder) URL(path string, args ...any) RequestBuilder {
	r.URL_ = fmt.Sprintf(path, args...)
	return r
}

func (r RequestBuilder) Query(key, value string, args ...any) RequestBuilder {
	r.Query_.Add(key, fmt.Sprintf(value, args...))
	return r
}

func (r RequestBuilder) GetURL() string {
	if r.Path_ != "" {
		panic("not supported")
	}

	return r.URL_
}

func (r RequestBuilder) Proxy(path string, args ...any) RequestBuilder {
	r.Proxy_ = fmt.Sprintf(path, args...)
	return r
}

func (r RequestBuilder) WithProxyTunnel() RequestBuilder {
	r.UseProxyTunnel_ = true
	return r
}

func (r RequestBuilder) DoNotFollowRedirects() RequestBuilder {
	r.DoNotFollowRedirects_ = true
	return r
}

func (r RequestBuilder) Method(method string) RequestBuilder {
	r.Method_ = method
	return r
}

func (r RequestBuilder) Header(k, v string) RequestBuilder {
	if r.Headers_ == nil {
		r.Headers_ = make([]HeaderBuilder, 0)
	}

	r.Headers_ = append(r.Headers_, Header(k, v))
	return r
}

func (r RequestBuilder) Headers(hs ...HeaderBuilder) RequestBuilder {
	if r.Headers_ == nil {
		r.Headers_ = make([]HeaderBuilder, 0)
	}

	r.Headers_ = append(r.Headers_, hs...)
	return r
}

type ExpectBuilder struct {
	StatusCode_ int             `json:"statusCode,omitempty"`
	Headers_    []HeaderBuilder `json:"headers,omitempty"`
	Body_       interface{}     `json:"body,omitempty"`
}

func Expect() ExpectBuilder {
	return ExpectBuilder{Body_: nil}
}

func (e ExpectBuilder) Status(statusCode int) ExpectBuilder {
	e.StatusCode_ = statusCode
	return e
}

func (e ExpectBuilder) Header(h HeaderBuilder) ExpectBuilder {
	e.Headers_ = append(e.Headers_, h)
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

func (e ExpectBuilder) Body(body interface{}) ExpectBuilder {
	switch body := body.(type) {
	case string:
		e.Body_ = []byte(body)
	case []byte:
		e.Body_ = body
	case check.Check[string]:
		e.Body_ = body
	case check.CheckWithHint[string]:
		e.Body_ = body
	default:
		panic("body must be string, []byte, or a regular check")
	}

	return e
}

func (e ExpectBuilder) BodyWithHint(hint string, body interface{}) ExpectBuilder {
	switch body := body.(type) {
	case string:
		e.Body_ = check.WithHint(
			hint,
			check.IsEqual(body),
		)
	case []byte:
		panic("body with hint for bytes is not implemented yet")
	case check.Check[string]:
		e.Body_ = check.WithHint(hint, body)
	case check.CheckWithHint[string]:
		panic("this check already has a hint")
	default:
		panic("body must be string, []byte, or a regular check")
	}

	return e
}

// TODO: split into two types: RequestHeaderBuidler and ResponseHeaderBuilder
type HeaderBuilder struct {
	Key_   string              `json:"key,omitempty"`
	Value_ Provider[string]    `json:"value,omitempty"`
	Check_ check.Check[string] `json:"-"`
	Hint_  string              `json:"-"`
}

func Header(key string, opts ...string) HeaderBuilder {
	if len(opts) > 1 {
		panic("too many options")
	}
	if len(opts) > 0 {
		return HeaderBuilder{Key_: key, Value_: StringProvider(opts[0]), Check_: check.IsEqual(opts[0])}
	}

	return HeaderBuilder{Key_: key}
}

func (h HeaderBuilder) Value(value string, args ...any) HeaderBuilder {
	h.Value_ = StringProvider(fmt.Sprintf(value, args...))
	return h
}

func (h HeaderBuilder) ValueFrom(value *string) HeaderBuilder {
	h.Value_ = FuncProvider[string](func() string {
		return *value
	})
	return h
}

func (h HeaderBuilder) ValueFromFunc(value func() string) HeaderBuilder {
	h.Value_ = FuncProvider[string](value)
	return h
}

func (h HeaderBuilder) Contains(value string, rest ...any) HeaderBuilder {
	h.Check_ = check.Contains(value, rest...)
	return h
}

func (h HeaderBuilder) Matches(value string, rest ...any) HeaderBuilder {
	h.Check_ = check.Matches(value, rest...)
	return h
}

func (h HeaderBuilder) Hint(hint string) HeaderBuilder {
	h.Hint_ = hint
	return h
}

func (h HeaderBuilder) Equals(value string, args ...any) HeaderBuilder {
	h.Check_ = check.IsEqual(value, args...)
	return h
}

func (h HeaderBuilder) IsEmpty() HeaderBuilder {
	h.Check_ = check.CheckIsEmpty{}
	return h
}

func (h HeaderBuilder) Checks(f func(string) bool) HeaderBuilder {
	h.Check_ = check.CheckFunc[string]{
		Fn: f,
	}
	return h
}
