package dnslink

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type DNSLinksAggregate struct {
	Domains map[string]string `json:"domains"`
}

func Aggregate(inputPaths []string) (*DNSLinksAggregate, error) {
	agg := DNSLinksAggregate{
		Domains: make(map[string]string),
	}

	for _, file := range inputPaths {
		dnsLinks, err := OpenDNSLink(file)
		if err != nil {
			return nil, fmt.Errorf("error loading file %s: %v", file, err)
		}

		for _, link := range dnsLinks.DNSLinks {
			if _, ok := agg.Domains[link.Domain]; ok {
				return nil, fmt.Errorf("collision detected for domain %s", link.Domain)
			}

			agg.Domains[link.Domain] = link.Path
			continue
		}
	}

	return &agg, nil
}

func MergeJSON(inputPaths []string, outputPath string) error {
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

// MergeEnv produces a string compatible with IPFS_NS_MAP env veriable syntax
// which can be used by tools to pre-populate namesys (IPNS, DNSLink) resolution
// results to facilitate tests based on static fixtures.
func MergeNsMapEnv(inputPaths []string, outputPath string) error {
	kvs, err := Aggregate(inputPaths)
	if err != nil {
		return err
	}

	var result []string
	for key, value := range kvs.Domains {
		result = append(result, fmt.Sprintf("%s:%s", key, value))
	}
	nsMapValue := strings.Join(result, ",")

	err = os.WriteFile(outputPath, []byte(nsMapValue), 0644)
	return err
}
