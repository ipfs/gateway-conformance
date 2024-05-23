package test

import (
	"net/url"
	"os"
	"strings"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("conformance")

func GetEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var GatewayURL = strings.TrimRight(
	GetEnv("GATEWAY_URL", "http://127.0.0.1:8080"),
	"/")

var SubdomainGatewayURL = strings.TrimRight(
	GetEnv("SUBDOMAIN_GATEWAY_URL", "http://example.com"),
	"/")

// EnableKuboLocalhostSubdomains is a flag that enables testing subdomains by querying no-port 'localhost'
// you can read more about why this is needed at https://github.com/ipfs/gateway-conformance/issues/185#issuecomment-2127598223
// default is true
var EnableKuboLocalhostSubdomains = GetEnv("ENABLE_KUBO_LOCALHOST_SUBDOMAINS", "true") == "true"

var GatewayHost = ""
var SubdomainGatewayHost = ""
var SubdomainGatewayScheme = ""

var SubdomainLocalhostGatewayURL = "http://localhost"

func init() {
	parsed, err := url.Parse(GatewayURL)
	if err != nil {
		panic(err)
	}

	GatewayHost = parsed.Host

	parsed, err = url.Parse(SubdomainGatewayURL)
	if err != nil {
		panic(err)
	}

	SubdomainGatewayHost = parsed.Host
	SubdomainGatewayScheme = parsed.Scheme

	log.Debugf("GatewayURL: %s", GatewayURL)

	log.Debugf("SubdomainGatewayURL: %s", SubdomainGatewayURL)
	log.Debugf("SubdomainGatewayHost: %s", SubdomainGatewayHost)
	log.Debugf("SubdomainGatewayScheme: %s", SubdomainGatewayScheme)
}
