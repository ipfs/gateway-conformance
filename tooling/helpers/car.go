package helpers

import (
	"fmt"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/test"
)

func StandardCARTestTransforms(t *testing.T, sts test.SugarTests) test.SugarTests {
	t.Helper()

	var out test.SugarTests
	for _, st := range sts {
		out = append(out, checkBothFormatURLParameterAndAcceptHeaderCAR(t, applyStandardCarResponseHeaders(t, st))...)
	}
	return out
}

// carResponseHeaders returns the standard headers expected for CAR responses
func carResponseHeaders() []test.HeaderBuilder {
	return []test.HeaderBuilder{
		// TODO: Go always sends Content-Length and it's not possible to explicitly disable the behavior.
		// For now, we ignore this check. It should be able to be resolved soon: https://github.com/ipfs/boxo/pull/177
		// test.Header("Content-Length").
		// 	Hint("CAR is streamed, gateway may not have the entire thing, unable to calculate total size").
		// 	IsEmpty(),
		test.Header("X-Content-Type-Options").
			Hint("Content type sniffing should be explicitly disabled via nosniff header.").
			Equals("nosniff"),
		test.Header("Accept-Ranges").
			Hint("CAR is streamed, gateway may not have the entire thing, unable to support range-requests. Partial downloads and resumes should be handled using IPLD selectors: https://github.com/ipfs/go-ipfs/issues/8769").
			Equals("none"),
		test.Header("Content-Type").
			Hint("Expected content type to be application/vnd.ipld.car").
			Contains("application/vnd.ipld.car"),
		test.Header("Content-Disposition").
			Hint(`Expected content disposition to be attachment; filename="*.car"`).
			Matches(`attachment; filename=".*\.car"`),
		test.Header("Etag").
			Hint("Etag must be present for caching purposes").
			Not().IsEmpty(),
	}
}

func applyStandardCarResponseHeaders(t *testing.T, st test.SugarTest) test.SugarTest {
	switch resp := st.Response.(type) {
	case test.AnyOfExpectBuilder:
		// Apply headers only to successful CAR responses (status 200)
		transformedExpects := make([]test.ExpectBuilder, 0, len(resp.Expect_))
		for _, expect := range resp.Expect_ {
			// Only apply CAR headers to 200 responses
			// 404/410 responses don't have CAR content, so no CAR headers needed
			if expect.StatusCode_ == 200 {
				expect = expect.Headers(carResponseHeaders()...)
			}
			transformedExpects = append(transformedExpects, expect)
		}
		st.Response = test.AnyOf(transformedExpects...)

	case test.ExpectBuilder:
		st.Response = resp.Headers(carResponseHeaders()...)

	default:
		t.Fatal("can only apply test transformation on an ExpectBuilder or AnyOfExpectBuilder")
	}

	return st
}

func checkBothFormatURLParameterAndAcceptHeaderCAR(t *testing.T, testWithFormatParam test.SugarTest) test.SugarTests {
	t.Helper()

	formatParamReq := testWithFormatParam.Request
	expected := testWithFormatParam.Response

	carFormatQueryParams, found := formatParamReq.Query_["format"]
	if !found {
		t.Fatal("could not find 'format' query parameter")
	}

	if len(carFormatQueryParams) != 1 {
		t.Fatal("only using a single format parameter is supported")
	}
	carFormatQueryParam := carFormatQueryParams[0]

	acceptHeaderReq := formatParamReq.Clone()
	delete(acceptHeaderReq.Query_, "format")

	return test.SugarTests{
		{
			Name:     fmt.Sprintf("%s (format=car)", testWithFormatParam.Name),
			Hint:     fmt.Sprintf("%s\n%s", testWithFormatParam.Hint, "Request using format=car"),
			Request:  formatParamReq,
			Response: expected,
		},
		{
			Name: fmt.Sprintf("%s (Accept Header)", testWithFormatParam.Name),
			Hint: fmt.Sprintf("%s\n%s", testWithFormatParam.Hint, "Request using an Accept header"),
			Request: acceptHeaderReq.
				Headers(
					test.Header("Accept", transformCARFormatParameterToAcceptHeader(t, carFormatQueryParam)),
				),
			Response: expected,
		},
	}
}

func transformCARFormatParameterToAcceptHeader(t *testing.T, param string) string {
	if param == "car" {
		return "application/vnd.ipld.car"
	}
	t.Fatalf("can only convert the CAR format parameter to an accept header. Got %q", param)
	return ""
}
