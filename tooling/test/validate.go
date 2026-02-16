package test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/check"
)

type testCheckOutput struct {
	testName    string
	specs       []string
	checkOutput check.CheckOutput
}

func validateResponse(
	t *testing.T,
	expected ExpectBuilder,
	res *http.Response,
) []testCheckOutput {
	t.Helper()

	var outputs []testCheckOutput

	if expected.StatusCode_ != 0 {
		output := testCheckOutput{testName: "Status code", checkOutput: check.CheckOutput{Success: true}}
		if res.StatusCode != expected.StatusCode_ {
			output.checkOutput.Success = false
			output.checkOutput.Reason = fmt.Sprintf("Status code is not %d. It is %d", expected.StatusCode_, res.StatusCode)
		}
		outputs = append(outputs, output)
	} else if expected.StatusCodeFrom_ != 0 && expected.StatusCodeTo_ != 0 {
		output := testCheckOutput{testName: "Status code", checkOutput: check.CheckOutput{Success: true}}
		if res.StatusCode < expected.StatusCodeFrom_ || res.StatusCode > expected.StatusCodeTo_ {
			output.checkOutput.Success = false
			output.checkOutput.Reason = fmt.Sprintf("Status code is not between %d and %d. It is %d", expected.StatusCodeFrom_, expected.StatusCodeTo_, res.StatusCode)
		}
	}

	for _, header := range expected.Headers_ {
		testName := fmt.Sprintf("Header %s", header.Key_)

		actual := res.Header.Values(header.Key_)

		// HTTP Headers can have multiple values, and that can be represented by comman separated value,
		// or sending the same header more than once. The `res.Header.Get` only returns the value
		// from the first header, so we use Values here.
		// At the same time, we don't want to have two separate checks everywhere, so we normalize
		// multiple instances of the same header by converting it into a single one, with comma-separated
		// values.
		if len(actual) > 1 {
			var result []string
			all := strings.Join(actual, ",")
			split := strings.SplitSeq(all, ",")
			for s := range split {
				value := strings.TrimSpace(s)
				if value != "" {
					result = append(result, strings.TrimSpace(s))
				}
			}
			// Normalize values from all instances of the header into a single one and comma-separated notation
			joined := strings.Join(result, ", ")
			actual = []string{joined}
		}

		c := header.Check_
		if header.Not_ {
			c = check.Not(c)
		}
		output := c.Check(actual)

		if !output.Success {
			if header.Hint_ == "" {
				output.Reason = fmt.Sprintf("Header '%s' %s", header.Key_, output.Reason)
			} else {
				output.Reason = fmt.Sprintf("Header '%s' %s (%s)", header.Key_, output.Reason, header.Hint_)
			}
		}

		outputs = append(outputs, testCheckOutput{testName: testName, checkOutput: output, specs: header.Specs_})
	}

	if expected.Body_ != nil {
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			outputs = append(outputs, testCheckOutput{testName: "Body", checkOutput: check.CheckOutput{Success: false, Reason: err.Error()}})
			return outputs
		}
		res.Body = io.NopCloser(bytes.NewBuffer(resBody))

		var output check.CheckOutput

		switch v := expected.Body_.(type) {
		case check.Check[string]:
			output = v.Check(string(resBody))
		case check.Check[[]byte]:
			output = v.Check(resBody)
		case string:
			output = check.IsEqual(v).Check(string(resBody))
		case []byte:
			output = check.IsEqualBytes(v).Check(resBody)
		default:
			output = check.CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("Body check has an invalid type: %T", expected.Body_),
			}
		}

		if !output.Success {
			if output.Hint == "" {
				output.Reason = fmt.Sprintf("Body %s", output.Reason)
			} else {
				output.Reason = fmt.Sprintf("Body %s (%s)", output.Reason, output.Hint)
			}
		}

		outputs = append(outputs, testCheckOutput{testName: "Body", checkOutput: output})
	}
	return outputs
}

func readPayload(res *http.Response) ([]byte, error) {
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

func validateResponses(
	t *testing.T,
	expected ExpectsBuilder,
	responses []*http.Response,
) {
	t.Helper()

	if expected.payloadsAreEquals {
		dumps := make([][]byte, 0, len(responses))

		for _, res := range responses {
			if res == nil {
				dumps = append(dumps, []byte("<nil>"))
			} else {
				// TODO: there is a usecase for mixing "one expect" (validate a single response)
				// and "expectS" (validates multiple responses). This will fail here if we check the body.
				// Support this usecase once this becomes a request feature.
				payload, err := readPayload(res)
				if err != nil {
					t.Errorf("Failed to read payload: %s", err)
				}
				dumps = append(dumps, payload)
			}
		}

		if len(dumps) > 1 {
			for i := 1; i < len(dumps); i++ {
				// if the payloads are not equal, we show an error
				if string(dumps[i]) != string(dumps[0]) {
					t.Errorf(`
Responses are not equal
==== Request %d ====

%s

==== Request %d ====

%s

`, 0+1, string(dumps[0]), i+1, string(dumps[i]))
				}
			}
		}
	}
}
