package tooling

import (
	"encoding/json"
	"testing"
)

func LogMetadata(t *testing.T, value any) {
	t.Helper()

	jsonValue, err := json.Marshal(value)
	if err != nil {
		t.Errorf("Failed to encode value: %v", err)
		return
	}
	t.Logf("--- META: %s", string(jsonValue))
}

func LogTestGroup(t *testing.T, name string) {
	t.Helper()

	LogMetadata(t, struct {
		Group string `json:"group"`
	}{
		Group: name,
	})
}

func LogVersion(t *testing.T) {
	LogMetadata(t, struct {
		Version string `json:"version"`
	}{
		Version: Version,
	})
}

func LogJobURL(t *testing.T) {
	LogMetadata(t, struct {
		JobURL string `json:"job_url"`
	}{
		JobURL: JobURL,
	})
}

func LogSpecs(t *testing.T, specs ...string) {
	if len(specs) == 0 {
		return
	}

	LogMetadata(t, struct {
		Specs []string `json:"specs"`
	}{
		Specs: specs,
	})
}
