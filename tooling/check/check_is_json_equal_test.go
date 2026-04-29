package check

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckIsJSONEqualMatch(t *testing.T) {
	c := IsJSONEqual([]byte(`{"a":1,"b":[1,2,3]}`))

	out := c.Check([]byte(`{"b":[1,2,3],"a":1}`))

	assert.True(t, out.Success, out.Reason)
}

func TestCheckIsJSONEqualMismatch(t *testing.T) {
	c := IsJSONEqual([]byte(`{"a":1}`))

	out := c.Check([]byte(`{"a":2}`))

	assert.False(t, out.Success)
	assert.Contains(t, out.Reason, "expected")
}

// Regression: gateway responses that are not JSON (e.g. an HTML wrapper page or
// upstream error) used to panic inside json.Unmarshal, taking the surrounding
// sub-test down with it. Check should report a graceful failure instead so
// callers can read the body from the failure reason and use Go's `-skip`
// against leaf sub-tests like .../Body without losing sibling coverage.
func TestCheckIsJSONEqualNonJSONBodyDoesNotPanic(t *testing.T) {
	c := IsJSONEqual([]byte(`{"hello":"world"}`))

	htmlBody := []byte(`<!doctype html><html><body>not json</body></html>`)

	out := c.Check(htmlBody)

	assert.False(t, out.Success)
	assert.Contains(t, out.Reason, "not valid JSON")
	assert.True(t, strings.Contains(out.Reason, "not json"), "reason should include the offending body for debugging, got: %s", out.Reason)
}

func TestCheckIsJSONEqualEmptyBodyDoesNotPanic(t *testing.T) {
	c := IsJSONEqual([]byte(`{}`))

	out := c.Check(nil)

	assert.False(t, out.Success)
	assert.Contains(t, out.Reason, "not valid JSON")
}
