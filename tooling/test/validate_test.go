package test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/stretchr/testify/assert"
)

// Regression: a gateway that returned a non-JSON body (e.g. an HTML wrapper
// page) for a test using check.IsJSONEqual used to panic inside
// json.Unmarshal, taking the parent sub-test down with it. validateResponse
// must surface the failure as a normal Body check output instead, so that
// consumers can use Go's `-skip` to skip the .../Body leaf without losing
// sibling Status_code / Header_* coverage.
//
// See https://github.com/ipfs/service-worker-gateway/pull/1039 for the
// downstream consumer that hit this.
func TestValidateResponseIsJSONEqualOnHTMLBodyDoesNotPanic(t *testing.T) {
	htmlBody := []byte(`<!doctype html><html><body>not json</body></html>`)

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/html"}},
		Body:       io.NopCloser(bytes.NewReader(htmlBody)),
	}

	expect := Expect().
		Status(200).
		Body(check.IsJSONEqual([]byte(`{"hello":"world"}`)))

	var outputs []testCheckOutput
	assert.NotPanics(t, func() {
		outputs = validateResponse(t, expect, res)
	})

	var bodyOutput *testCheckOutput
	for i := range outputs {
		if outputs[i].testName == "Body" {
			bodyOutput = &outputs[i]
			break
		}
	}

	assert.NotNil(t, bodyOutput, "expected a Body check output")
	if bodyOutput != nil {
		assert.False(t, bodyOutput.checkOutput.Success, "Body check must fail on non-JSON body")
		assert.Contains(t, bodyOutput.checkOutput.Reason, "not valid JSON")
	}
}
