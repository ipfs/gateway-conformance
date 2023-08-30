package test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/specs"
)

type SugarTest struct {
	Name      string
	Hint      string
	Spec      string
	Specs     []string
	Request   RequestBuilder
	Requests  []RequestBuilder
	Response  ExpectValidator
	Responses ExpectsBuilder
}

type SugarTests []SugarTest

func (s *SugarTest) AllSpecs() []string {
	if len(s.Specs) > 0 && s.Spec != "" {
		panic("cannot have both Spec and Specs")
	}

	if len(s.Specs) > 0 {
		return s.Specs
	}

	if s.Spec != "" {
		return []string{s.Spec}
	}

	return []string{}
}

func RunWithSpecs(
	t *testing.T,
	tests SugarTests,
	required ...specs.Leaf,
) {
	t.Helper()

	missing := []specs.Spec{}
	for _, spec := range required {
		if !spec.IsEnabled() {
			missing = append(missing, spec)
		}
	}

	if len(missing) > 0 {
		t.Skipf("skipping tests, missing specs: %v", missing)
		return
	}

	run(t, tests)
}

func run(t *testing.T, tests SugarTests) {
	t.Helper()

	for _, test := range tests {
		timeout, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		name := safeName(test.Name)

		if len(test.Requests) > 0 {
			t.Run(name, func(t *testing.T) {
				tooling.LogSpecs(t, test.AllSpecs()...)
				responses := make([]*http.Response, 0, len(test.Requests))

				for _, req := range test.Requests {
					_, res, localReport := runRequest(timeout, t, test, req)
					if test.Response != nil {
						test.Response.Validate(t, res, localReport)
					}
					responses = append(responses, res)
				}

				validateResponses(t, test.Responses, responses)
			})
		} else {
			t.Run(name, func(t *testing.T) {
				tooling.LogSpecs(t, test.AllSpecs()...)
				_, res, localReport := runRequest(timeout, t, test, test.Request)
				if test.Response != nil {
					test.Response.Validate(t, res, localReport)
				}
			})
		}
	}
}

func safeName(s string) string {
	// Split the string by spaces
	parts := strings.Split(s, " ")

	// Escape each part
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}

	// Join the parts back together with spaces
	return strings.Join(parts, " ")
}
