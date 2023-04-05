package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ipfs/gateway-conformance/tooling/dnslink"
	"github.com/ipfs/gateway-conformance/tooling/fixtures"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: dnslinkgen <domain>")
		os.Exit(1)
	}

	domain := os.Args[1]

	fxs, err := fixtures.List()
	if err != nil {
		log.Fatal(err)
	}

	configs := fxs.ConfigFiles
	aggMap, err := dnslink.Aggregate(configs)
	if err != nil {
		log.Fatal(err)
	}

	// print k=v on stdout
	var kvs []string
	for k, v := range aggMap {
		kvs = append(kvs, fmt.Sprintf("%s%s:%s", k, domain, v))
	}

	fmt.Println("export IPFS_NS_MAP=\"" + strings.Join(kvs, ",") + "\"")
}
