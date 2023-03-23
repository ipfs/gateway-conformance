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

var GatewayUrl = strings.TrimRight(
	GetEnv("GATEWAY_URL", "http://127.0.0.1:8080"),
	"/")

var SubdomainGatewayUrl = strings.TrimRight(
	GetEnv("SUBDOMAIN_GATEWAY_URL", "http://example.com"),
	"/")


var GatewayHost = ""
var SubdomainGatewayHost = ""
var SubdomainGatewayScheme = ""

var SubdomainLocalhostGatewayUrl = "http://localhost"

func init() {
	parsed, err := url.Parse(GatewayUrl)
	if err != nil {
		panic(err)
	}

	GatewayHost = parsed.Host

	parsed, err = url.Parse(SubdomainGatewayUrl)
	if err != nil {
		panic(err)
	}

	SubdomainGatewayHost = parsed.Host
	SubdomainGatewayScheme = parsed.Scheme

	log.Debugf("GatewayUrl: %s", GatewayUrl)

	log.Debugf("SubdomainGatewayUrl: %s", SubdomainGatewayUrl)
	log.Debugf("SubdomainGatewayHost: %s", SubdomainGatewayHost)
	log.Debugf("SubdomainGatewayScheme: %s", SubdomainGatewayScheme)
}