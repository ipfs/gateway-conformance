package tooling

import (
	"encoding/json"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/test"
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

func LogGatewayURL(t *testing.T) {
	LogMetadata(t, struct {
		GatewayURL          string `json:"gateway_url"`
		SubdomainGatewayURL string `json:"subdomain_gateway_url"`
	}{
		GatewayURL:          test.GatewayURL,
		SubdomainGatewayURL: test.SubdomainGatewayURL,
	})
}
