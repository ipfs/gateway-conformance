#!/usr/bin/env bash

FIXTURES_PATH=${1:-$(pwd)}

ipfs config --json Gateway.PublicGateways '{
	"example.com": {
		"UseSubdomains": true,
		"InlineDNSLink": true,
		"Paths": ["/ipfs", "/ipns"]
	},
	"localhost": {
		"UseSubdomains": true,
		"InlineDNSLink": true,
		"Paths": ["/ipfs", "/ipns"]
	}
}'

export IPFS_NS_MAP="$(cat "${FIXTURES_PATH}/dnslinks.IPFS_NS_MAP")"

echo "Set the following IPFS_NS_MAP before starting the kubo daemon:"
echo "IPFS_NS_MAP=${IPFS_NS_MAP}"
