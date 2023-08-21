package test

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/check"
)

func validateResponse(
	t *testing.T,
	expected ExpectBuilder,
	res *http.Response,
	localReport Reporter,
) {
	t.Helper()

	if expected.StatusCode_ != 0 {
		if res.StatusCode != expected.StatusCode_ {
			localReport(t, "Status code is not %d. It is %d", expected.StatusCode_, res.StatusCode)
		}
	}

	for _, header := range expected.Headers_ {
		t.Run(fmt.Sprintf("Header %s", header.Key_), func(t *testing.T) {
			MustNotBeSkipped(t)
			actual := res.Header.Values(header.Key_)

			c := header.Check_
			if header.Not_ {
				c = check.Not(c)
			}
			output := c.Check(actual)

			if !output.Success {
				if header.Hint_ == "" {
					localReport(t, "Header '%s' %s", header.Key_, output.Reason)
				} else {
					localReport(t, "Header '%s' %s (%s)", header.Key_, output.Reason, header.Hint_)
				}
			}
		})
	}

	if expected.Body_ != nil {
		t.Run("Body", func(t *testing.T) {
			MustNotBeSkipped(t)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				localReport(t, err)
			}

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
					localReport(t, "Body %s", output.Reason)
				} else {
					localReport(t, "Body %s (%s)", output.Reason, output.Hint)
				}
			}
		})
	}
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
