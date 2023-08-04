package tooling

import (
	"encoding/json"
	"testing"
)

func LogMetadata(t *testing.T, value interface{}) {
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

func LogIPIP(t *testing.T, ipip string, sections ...string) {
	t.Helper()

	LogMetadata(t, struct {
		IPIP     string   `json:"ipip"`
		Sections []string `json:"sections,omitempty"`
	}{
		IPIP:     ipip,
		Sections: sections,
	})
}

func LogVersion(t *testing.T) {
	LogMetadata(t, struct {
		Version string `json:"version"`
	}{
		Version: Version,
	})
}
