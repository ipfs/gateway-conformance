package dnslink

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ipfs/gateway-conformance/tooling/fixtures"
	"gopkg.in/yaml.v3"
)

type ConfigFixture struct {
	DNSLinks map[string]DNSLink `yaml:"dnslinks"`
}

type DNSLink struct {
	Subdomain string `yaml:"subdomain"`
	Domain    string `yaml:"domain"`
	Path      string `yaml:"path"`
}

func InlineDNS(s string) string {
	// See spec at https://github.com/ipfs/specs/blob/main/src/http-gateways/subdomain-gateway.md#host-request-header
	// Every - is replaced with --
	// Every . is replaced with -
	return strings.ReplaceAll(strings.ReplaceAll(s, "-", "--"), ".", "-")
}

func OpenDNSLink(absPath string) (*ConfigFixture, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var dnsLinks ConfigFixture
	err = yaml.Unmarshal(data, &dnsLinks)
	if err != nil {
		return nil, err
	}

	return &dnsLinks, nil
}

func MustOpenDNSLink(file string) *ConfigFixture {
	fixturePath := path.Join(fixtures.Dir(), file)
	dnsLinks, err := OpenDNSLink(fixturePath)
	if err != nil {
		panic(err)
	}

	return dnsLinks
}

func (d *ConfigFixture) MustGet(id string) string {
	dnsLink, ok := d.DNSLinks[id]
	if !ok {
		panic(fmt.Errorf("dnslink %s not found", id))
	}
	if dnsLink.Domain != "" && dnsLink.Subdomain != "" {
		panic(fmt.Errorf("dnslink %s has both domain and subdomain", id))
	}
	if dnsLink.Domain == "" && dnsLink.Subdomain == "" {
		panic(fmt.Errorf("dnslink %s has neither domain nor subdomain", id))
	}
	if dnsLink.Path == "" {
		panic(fmt.Errorf("dnslink %s has no path", id))
	}

	if dnsLink.Domain != "" {
		return dnsLink.Domain
	}

	return dnsLink.Subdomain
}
