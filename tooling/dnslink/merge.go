package dnslink

import (
	"encoding/json"
	"fmt"
	"os"
)

type DNSLinksAggregate struct {
	Domains    map[string]string `json:"domains"`
	Subdomains map[string]string `json:"subdomains"`
}

func Aggregate(inputPaths []string) (*DNSLinksAggregate, error) {
	agg := DNSLinksAggregate{
		Domains:    make(map[string]string),
		Subdomains: make(map[string]string),
	}

	for _, file := range inputPaths {
		dnsLinks, err := OpenDNSLink(file)
		if err != nil {
			return nil, fmt.Errorf("error loading file %s: %v", file, err)
		}

		for _, link := range dnsLinks.DNSLinks {
			if link.Domain != "" && link.Subdomain != "" {
				return nil, fmt.Errorf("dnslink %s has both domain and subdomain", link.Subdomain)
			}

			if link.Domain != "" {
				if _, ok := agg.Domains[link.Domain]; ok {
					return nil, fmt.Errorf("collision detected for domain %s", link.Domain)
				}

				agg.Domains[link.Domain] = link.Path
				continue
			}

			if link.Subdomain != "" {
				if _, ok := agg.Subdomains[link.Subdomain]; ok {
					return nil, fmt.Errorf("collision detected for subdomain %s", link.Subdomain)
				}

				agg.Subdomains[link.Subdomain] = link.Path
				continue
			}
		}
	}

	return &agg, nil
}

func Merge(inputPaths []string, outputPath string) error {
	kvs, err := Aggregate(inputPaths)
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(kvs, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(outputPath, j, 0644)
	return err
}
