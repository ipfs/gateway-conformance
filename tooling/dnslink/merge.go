package dnslink

import (
	"encoding/json"
	"fmt"
	"os"
)

func Aggregate(inputPaths []string) (map[string]string, error) {
	aggMap := make(map[string]string)

	for _, file := range inputPaths {
		dnsLinks, err := OpenDNSLink(file)
		if err != nil {
			return nil, fmt.Errorf("error loading file %s: %v", file, err)
		}

		for _, link := range dnsLinks.DNSLinks {
			if _, ok := aggMap[link.Subdomain]; ok {
				return nil, fmt.Errorf("collision detected for subdomain %s", link.Subdomain)
			}

			aggMap[link.Subdomain] = link.Path
		}
	}

	return aggMap, nil
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
