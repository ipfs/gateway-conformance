#! /usr/bin/env bash
ipfs config --json Gateway.PublicGateways '{	
	"example.com": {							
		"UseSubdomains": true,			 		
		"Paths": ["/ipfs", "/ipns", "/api"]		
	},											
	"localhost": {								
		"UseSubdomains": true,					
		"InlineDNSLink": true,					
		"Paths": ["/ipfs", "/ipns", "/api"]		
	}											
}'

IPFS_NS_MAP=$(cat ./dnslinks.json | jq -r 'to_entries | map("\(.key).example.com:\(.value)") | join(",")')

echo "Set the following IPFS_NS_MAP before starting the kubo daemon:"
echo "IPFS_NS_MAP=${IPFS_NS_MAP}"
