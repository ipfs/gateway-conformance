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
