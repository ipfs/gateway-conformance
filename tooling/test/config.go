package test

import (
	"context"
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
	GetEnv("GATEWAY_URL", "http://localhost:8080"),
	"/")

var SubdomainGatewayUrl = strings.TrimRight(
	GetEnv("SUBDOMAIN_GATEWAY_URL", "http://example.com:8080"),
	"/")

// This domain is used as a placeholder,
// A test implementer would use `example.com` to write an explicit test.
// At test time, we replace this with the actual domain configured by the test runner.
const GATEWAY_EXAMPLE_DOMAIN = "example.com"

var GatewayHost = ""
var SubdomainGatewayHost = ""
var SubdomainGatewayScheme = ""

func init() {
	parse, err := url.Parse(GatewayUrl)
	if err != nil {
		panic(err)
	}

	GatewayHost = parse.Host

	parse, err = url.Parse(SubdomainGatewayUrl)
	if err != nil {
		panic(err)
	}

	SubdomainGatewayHost = parse.Host
	SubdomainGatewayScheme = parse.Scheme

	log.Debugf("SubdomainGatewayHost: %s", SubdomainGatewayHost)
	log.Debugf("SubdomainGatewayScheme: %s", SubdomainGatewayScheme)
}

func NewDialer() *net.Dialer {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				log.Debugf("Custom Resolver dialing into network: %s, address: %s", network, address)

				d := net.Dialer{
					Timeout: 30 * time.Second,
				}

				return d.DialContext(ctx, network, address)
			},
		},
	}

	http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		log.Debugf("Custom Dialer dialing into network: %s, address: %s", network, addr)

		// If we call into a subdomain `somethingsomething.example.com`,
		// actually dial the gateway url on its base address (probably localhost:8080)
		if strings.HasSuffix(addr, SubdomainGatewayHost) {
			addr = GatewayHost
		}

		log.Debugf("Custom Dialer dialing into (effective) network: %s, address: %s", network, addr)
		conn, err := dialer.DialContext(ctx, network, addr)
		return conn, err
	}

	return dialer
}
