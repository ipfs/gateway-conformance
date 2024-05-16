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
