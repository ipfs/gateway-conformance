package check

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckHas(t *testing.T) {
	// Create a Check instance
	check := Has("value1", "value2", "value3")

	// Test data
	testData := []string{"value1", "value2", "value3"}

	// Call .Check on the instance
	checkOutput := check.Check(testData)

	assert.True(t, checkOutput.Success, checkOutput.Reason)
}

func TestCheckHasFailure(t *testing.T) {
	// Create a Check instance
	check := Has("value3")

	// Test data
	testData := []string{"value1", "value2"}

	// Call .Check on the instance
	checkOutput := check.Check(testData)

	assert.False(t, checkOutput.Success, checkOutput.Reason)
}
