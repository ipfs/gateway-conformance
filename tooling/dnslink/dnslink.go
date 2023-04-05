package dnslink

import (
	"fmt"
	"os"
	"path"

	"github.com/ipfs/gateway-conformance/tooling/fixtures"
	"gopkg.in/yaml.v3"
)

type DNSLinks struct {
	DNSLinks map[string]DNSLink `yaml:"dnslinks"`
}

type DNSLink struct {
	Subdomain string `yaml:"subdomain"`
	Path      string `yaml:"path"`
}

func OpenDNSLink(absPath string) (*DNSLinks, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var dnsLinks DNSLinks
	err = yaml.Unmarshal(data, &dnsLinks)
	if err != nil {
		return nil, err
	}

	return &dnsLinks, nil
}

func MustOpenDNSLink(file string) *DNSLinks {
	fixturePath := path.Join(fixtures.Dir(), file)
	dnsLinks, err := OpenDNSLink(fixturePath)
	if err != nil {
		panic(err)
	}

	return dnsLinks
}

func (d *DNSLinks) Get(id string) string {
	dnsLink, ok := d.DNSLinks[id]
	if !ok {
		panic(fmt.Errorf("dnslink %s not found", id))
	}
	return dnsLink.Subdomain
}
