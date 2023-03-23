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