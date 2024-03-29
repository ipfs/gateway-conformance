#! /usr/bin/env bash

FIXTURES_PATH=${1:-$(pwd)}

ipfs config --json Gateway.PublicGateways '{
	"example.com": {
		"UseSubdomains": true,
		"InlineDNSLink": true,
		"Paths": ["/ipfs", "/ipns", "/api"]
	},
	"localhost": {
		"UseSubdomains": true,
		"InlineDNSLink": true,
		"Paths": ["/ipfs", "/ipns", "/api"]
	}
}'

export IPFS_NS_MAP=$(cat "${FIXTURES_PATH}/dnslinks.json" | jq -r '.subdomains | to_entries | map("\(.key).example.com:\(.value)") | join(",")')
export IPFS_NS_MAP="$(cat "${FIXTURES_PATH}/dnslinks.json" | jq -r '.domains | to_entries | map("\(.key):\(.value)") | join(",")'),${IPFS_NS_MAP}"

echo "Set the following IPFS_NS_MAP before starting the kubo daemon:"
echo "IPFS_NS_MAP=${IPFS_NS_MAP}"
