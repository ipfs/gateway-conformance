package test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

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

const GATEWAY_LOCALHOST_DOMAIN = "localhost"
var SubdomainLocalhostGatewayUrl = ""

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

	// Generate the localhost subdomain gateway url
	parsed.Scheme = "http"
	if parsed.Port() == "" {
		parsed.Host = "localhost"
	} else {
		parsed.Host = fmt.Sprintf("localhost:%s", parsed.Port())
	}
	SubdomainLocalhostGatewayUrl = parsed.String()
}

func NewDialer() *net.Dialer {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				// log.Debugf("Custom Resolver dialing into network: %s, address: %s", network, address)

				d := net.Dialer{
					Timeout: 30 * time.Second,
				}

				return d.DialContext(ctx, network, address)
			},
		},
	}

	http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		effectiveAddr := addr

		// If we call into a subdomain `somethingsomething.example.com`,
		// actually dial the gateway url on its base address (probably localhost:8080)
		if strings.HasSuffix(addr, SubdomainGatewayHost) {
			effectiveAddr = GatewayHost
		}

		// If we call into a subdomain `somethingsomething.localhost`,
		// actually dial the gateway url on its base address (probably localhost:8080)
		if strings.HasSuffix(addr, GATEWAY_LOCALHOST_DOMAIN) || strings.HasSuffix(addr, fmt.Sprintf("%s:80", GATEWAY_LOCALHOST_DOMAIN)) {
			effectiveAddr = GatewayHost
		}

		if effectiveAddr != addr {
			log.Debugf("Custom Dialer dialing and rewrote (network: %s) %s => %s", network, addr, effectiveAddr)
		} else {
			log.Debugf("Custom Dialer dialing (network: %s) %s", network, effectiveAddr)
		}

		conn, err := dialer.DialContext(ctx, network, effectiveAddr)
		return conn, err
	}

	return dialer
}
