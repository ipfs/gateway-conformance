package test

import (
	"net/url"
	"os"
	"strings"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("conformance")

func env2url(key string) *url.URL {
	value, ok := os.LookupEnv(key)
	if !ok {
		// Actual validation of GATEWAY_URL and SUBDOMAIN_GATEWAY_URL happens in
		// in cmd/gateway-conformance/main.go. Here we create a bogus URL
		// to ensure no error at compilation time.
		value = "http://todo-changeme.example.com"
	}
	gatewayURL := strings.TrimRight(value, "/")
	parsed, err := url.Parse(gatewayURL)
	if err != nil {
		panic(err)
	}
	return parsed
}

func GatewayURL() *url.URL {
	return env2url("GATEWAY_URL")
}

func SubdomainGatewayURL() *url.URL {
	return env2url("SUBDOMAIN_GATEWAY_URL")
}
