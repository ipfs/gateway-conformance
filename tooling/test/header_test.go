package test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
	// Manually create a response object with headers
	resp := &http.Response{
		Header: http.Header{
			"Access-Control-Allow-Headers": []string{"Content-Type, Range, User-Agent, X-Requested-With"},
		},
	}

	// Extract headers from the response
	headers := resp.Header["Access-Control-Allow-Headers"]

	// Test the HeaderBuilder function
	hb := Header("Access-Control-Allow-Headers").Has("Content-Type", "Range", "User-Agent", "X-Requested-With")

	checkOutput := hb.Check_.Check(headers)

	// Check if the headers satisfy the conditions
	assert.True(t, checkOutput.Success, checkOutput.Reason)
}

func TestHeaderBuilderClone(t *testing.T) {
	hb := Header("Etag", `"someCID123"`)

	cloned := hb.Clone()

	assert.Equal(t, hb.Key_, cloned.Key_, "Clone must preserve Key_")
	assert.Equal(t, hb.Value_, cloned.Value_, "Clone must preserve Value_")
	assert.Equal(t, hb.Hint_, cloned.Hint_, "Clone must preserve Hint_")
	assert.Equal(t, hb.Not_, cloned.Not_, "Clone must preserve Not_")
	assert.NotEqual(t, cloned.Value_, cloned.Key_, "Value_ must not be copied from Key_")
}

func TestHeaderFailure(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{
			// missing X-Requested-With
			"Access-Control-Allow-Headers": []string{"Content-Type, Range, User-Agent"},
		},
	}
	// Extract headers from the response
	headers := resp.Header["Access-Control-Allow-Headers"]

	// Test the HeaderBuilder function
	hb := Header("Access-Control-Allow-Headers").Has("Content-Type", "Range", "User-Agent", "X-Requested-With")

	checkOutput := hb.Check_.Check(headers)

	assert.False(t, checkOutput.Success, checkOutput.Reason)
}
