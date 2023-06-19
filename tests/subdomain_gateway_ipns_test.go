package tests

import (
	"net/url"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	"github.com/ipfs/gateway-conformance/tooling/dnslink"
	"github.com/ipfs/gateway-conformance/tooling/helpers"
	"github.com/ipfs/gateway-conformance/tooling/ipns"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multicodec"
)

// TODO(laurent): this were in t0114_gateway_subdomains_test.go

func TestGatewaySubdomainAndIPNS(t *testing.T) {
	tests := SugarTests{}

	rsaFixture := ipns.MustOpenIPNSRecordWithKey("t0114/QmVujd5Vb7moysJj8itnGufN7MEtPRCNHkKpNuA4onsRa3.ipns-record")
	ed25519Fixture := ipns.MustOpenIPNSRecordWithKey("t0114/12D3KooWLQzUv2FHWGVPXTXSZpdHs7oHbXub2G5WC8Tx4NQhyd2d.ipns-record")

	car := car.MustOpenUnixfsCar("t0114/fixtures.car")
	helloCID := "bafkreicysg23kiwv34eg2d7qweipxwosdo2py4ldv42nbauguluen5v6am"
	payload := string(car.MustGetRawData(helloCID))

	// We're going to run the same test against multiple gateways (localhost, and a subdomain gateway)
	gatewayURLs := []string{
		SubdomainGatewayURL,
		SubdomainLocalhostGatewayURL,
	}

	ipnsRecords := []*ipns.IpnsRecord{
		rsaFixture,
		ed25519Fixture,
	}

	for _, gatewayURL := range gatewayURLs {
		u, err := url.Parse(gatewayURL)
		if err != nil {
			t.Fatal(err)
		}

		for _, record := range ipnsRecords {
			tests = append(tests, SugarTests{
				// # /ipns/<libp2p-key>
				// test_localhost_gateway_response_should_contain \
				//   "request for localhost/ipns/{CIDv0} redirects to CIDv1 with libp2p-key multicodec in subdomain" \
				//   "http://localhost:$GWAY_PORT/ipns/$RSA_IPNS_IDv0" \
				//   "Location: http://${RSA_IPNS_IDv1}.ipns.localhost:$GWAY_PORT/"
				// test_localhost_gateway_response_should_contain \
				//   "request for localhost/ipns/{CIDv0} redirects to CIDv1 with libp2p-key multicodec in subdomain" \
				//   "http://localhost:$GWAY_PORT/ipns/$ED25519_IPNS_IDv0" \
				//   "Location: http://${ED25519_IPNS_IDv1}.ipns.localhost:$GWAY_PORT/"
				{
					Name: "request for /ipns/{CIDv0} redirects to CIDv1 with libp2p-key multicodec in subdomain",
					Request: Request().
						URL("{{url}}/ipns/{{cid}}", gatewayURL, record.IdV0()),
					Response: Expect().
						Status(301).
						Headers(
							Header("Location").
								Equals("{{scheme}}://{{cid}}.ipns.{{host}}/", u.Scheme, record.IdV1(), u.Host),
						),
				},
				// # *.ipns.localhost
				// # <libp2p-key>.ipns.localhost
				// test_localhost_gateway_response_should_contain \
				//   "request for {CIDv1-libp2p-key}.ipns.localhost returns expected payload" \
				//   "http://${RSA_IPNS_IDv1}.ipns.localhost:$GWAY_PORT" \
				//   "$CID_VAL"
				// test_localhost_gateway_response_should_contain \
				//   "request for {CIDv1-libp2p-key}.ipns.localhost returns expected payload" \
				//   "http://${ED25519_IPNS_IDv1}.ipns.localhost:$GWAY_PORT" \
				//   "$CID_VAL"
				{
					Name: "request for {CIDv1-libp2p-key}.ipns.{gateway} returns expected payload",
					Request: Request().
						URL("{{scheme}}://{{cid}}.ipns.{{host}}/", u.Scheme, record.IdV1(), u.Host),
					Response: Expect().
						Status(200).
						BodyWithHint("Request for {{cid}}.ipns.{{host}} returns expected payload", payload),
				},
				// test_localhost_gateway_response_should_contain \
				//   "localhost request for {CIDv1-dag-pb}.ipns.localhost redirects to CID with libp2p-key multicodec" \
				//   "http://${RSA_IPNS_IDv1_DAGPB}.ipns.localhost:$GWAY_PORT" \
				//   "Location: http://${RSA_IPNS_IDv1}.ipns.localhost:$GWAY_PORT/"
				// test_localhost_gateway_response_should_contain \
				//   "localhost request for {CIDv1-dag-pb}.ipns.localhost redirects to CID with libp2p-key multicodec" \
				//   "http://${ED25519_IPNS_IDv1_DAGPB}.ipns.localhost:$GWAY_PORT" \
				//   "Location: http://${ED25519_IPNS_IDv1}.ipns.localhost:$GWAY_PORT/"
				{
					Name: "request for {CIDv1-dag-pb}.ipns.{gateway} redirects to CID with libp2p-key multicodec",
					Request: Request().
						URL("{{scheme}}://{{cid}}.ipns.{{host}}/", u.Scheme, record.ToCID(multicodec.DagPb, multibase.Base36), u.Host),
					Response: Expect().
						Status(301).
						Headers(
							Header("Location").
								Equals("{{scheme}}://{{cid}}.ipns.{{host}}/", u.Scheme, record.IdV1(), u.Host),
						),
				},
				// # example.com/ipns/<libp2p-key>
				// test_hostname_gateway_response_should_contain \
				//   "request for example.com/ipns/{CIDv0} redirects to CIDv1 with libp2p-key multicodec in subdomain" \
				//   "example.com" \
				//   "http://127.0.0.1:$GWAY_PORT/ipns/$RSA_IPNS_IDv0" \
				//   "Location: http://${RSA_IPNS_IDv1}.ipns.example.com/"
				// test_hostname_gateway_response_should_contain \
				//   "request for example.com/ipns/{CIDv0} redirects to CIDv1 with libp2p-key multicodec in subdomain" \
				//   "example.com" \
				//   "http://127.0.0.1:$GWAY_PORT/ipns/$ED25519_IPNS_IDv0" \
				//   "Location: http://${ED25519_IPNS_IDv1}.ipns.example.com/"
				// NOTE: Done above, thanks to the loop
				//
				// # *.ipns.example.com
				// # ============================================================================

				// # <libp2p-key>.ipns.example.com

				// test_hostname_gateway_response_should_contain \
				//   "request for {CIDv1-libp2p-key}.ipns.example.com returns expected payload" \
				//   "${RSA_IPNS_IDv1}.ipns.example.com" \
				//   "http://127.0.0.1:$GWAY_PORT" \
				//   "$CID_VAL"

				// test_hostname_gateway_response_should_contain \
				//   "request for {CIDv1-libp2p-key}.ipns.example.com returns expected payload" \
				//   "${ED25519_IPNS_IDv1}.ipns.example.com" \
				//   "http://127.0.0.1:$GWAY_PORT" \
				//   "$CID_VAL"

				// test_hostname_gateway_response_should_contain \
				//   "hostname request for {CIDv1-dag-pb}.ipns.localhost redirects to CID with libp2p-key multicodec" \
				//   "${RSA_IPNS_IDv1_DAGPB}.ipns.example.com" \
				//   "http://127.0.0.1:$GWAY_PORT" \
				//   "Location: http://${RSA_IPNS_IDv1}.ipns.example.com/"

				// test_hostname_gateway_response_should_contain \
				//   "hostname request for {CIDv1-dag-pb}.ipns.localhost redirects to CID with libp2p-key multicodec" \
				//   "${ED25519_IPNS_IDv1_DAGPB}.ipns.example.com" \
				//   "http://127.0.0.1:$GWAY_PORT" \
				//   "Location: http://${ED25519_IPNS_IDv1}.ipns.example.com/"
				// # disable /ipns for the hostname by not whitelisting it
				// ipfs config --json Gateway.PublicGateways '{
				//   "example.com": {
				//     "UseSubdomains": true,
				//     "Paths": ["/ipfs"]
				//   }
				// }' || exit 1
				// # restart daemon to apply config changes
				// test_kill_ipfs_daemon
				// test_launch_ipfs_daemon_without_network

				// TODO: what to do with these?
				// # refuse requests to Paths that were not explicitly whitelisted for the hostname
				// test_hostname_gateway_response_should_contain \
				//   "request for *.ipns.example.com returns HTTP 404 Not Found when /ipns is not on Paths whitelist" \
				//   "${RSA_IPNS_IDv1}.ipns.example.com" \
				//   "http://127.0.0.1:$GWAY_PORT" \
				//   "404 Not Found"

				// test_hostname_gateway_response_should_contain \
				//   "request for *.ipns.example.com returns HTTP 404 Not Found when /ipns is not on Paths whitelist" \
				//   "${ED25519_IPNS_IDv1}.ipns.example.com" \
				//   "http://127.0.0.1:$GWAY_PORT" \
				//   "404 Not Found"

				// # refuse requests to Paths that were not explicitly whitelisted for the hostname
				// test_hostname_gateway_response_should_contain \
				//   "request for example.com/ipns/ returns HTTP 404 Not Found when /ipns is not on Paths whitelist" \
				//   "example.com" \
				//   "http://127.0.0.1:$GWAY_PORT/ipns/$RSA_IPNS_IDv1" \
				//   "404 Not Found"

				// test_hostname_gateway_response_should_contain \
				//   "request for example.com/ipns/ returns HTTP 404 Not Found when /ipns is not on Paths whitelist" \
				//   "example.com" \
				//   "http://127.0.0.1:$GWAY_PORT/ipns/$ED25519_IPNS_IDv1" \
				//   "404 Not Found"
			}...)
		}

		tests = append(tests, SugarTests{
			// ## Test subdomain handling of CIDs that do not fit in a single DNS Label (>63chars)
			// ## https://github.com/ipfs/go-ipfs/issues/7318
			// ## ============================================================================
			// # local: *.localhost
			// test_localhost_gateway_response_should_contain \
			//   "request for a ED25519 libp2p-key at localhost/ipns/{b58mh} returns Location HTTP header for DNS-safe subdomain redirect in browsers" \
			//   "http://localhost:$GWAY_PORT/ipns/$IPNS_ED25519_B58MH" \
			//   "Location: http://${IPNS_ED25519_B36CID}.ipns.localhost:$GWAY_PORT/"
			// # public subdomain gateway: *.example.com
			// test_hostname_gateway_response_should_contain \
			//   "request for a ED25519 libp2p-key at example.com/ipns/{b58mh} returns Location HTTP header for DNS-safe subdomain redirect in browsers" \
			//   "example.com" \
			//   "http://127.0.0.1:$GWAY_PORT/ipns/$IPNS_ED25519_B58MH" \
			//   "Location: http://${IPNS_ED25519_B36CID}.ipns.example.com"
			{
				Name: "request for a ED25519 libp2p-key at example.com/ipns/{b58mh} returns Location HTTP header for DNS-safe subdomain redirect in browsers",
				Request: Request().
					URL("{{url}}/ipns/{{cid}}", gatewayURL, ed25519Fixture.B58MH()),
				Response: Expect().
					Headers(
						Header("Location").
							Equals("{{scheme}}://{{cid}}.ipns.{{host}}/", u.Scheme, ed25519Fixture.ToCID(multicodec.Libp2pKey, multibase.Base36), u.Host),
					),
			},
		}...)

	}

	RunWithSpecs(t, helpers.UnwrapSubdomainTests(t, tests), specs.SubdomainGatewayIPNS)
}

// TODO(laurent): this were in t0114_gateway_subdomains_test.go

func TestSubdomainGatewayDNSLinkInlining(t *testing.T) {
	tests := SugarTests{}

	// We're going to run the same test against multiple gateways (localhost, and a subdomain gateway)
	gatewayURLs := []string{
		SubdomainGatewayURL,
		SubdomainLocalhostGatewayURL,
	}

	dnsLinks := dnslink.MustOpenDNSLink("t0114/dnslink.yml")
	wikipedia := dnsLinks.MustGet("wikipedia")
	dnsLinkTest := dnsLinks.MustGet("test")

	for _, gatewayURL := range gatewayURLs {
		u, err := url.Parse(gatewayURL)
		if err != nil {
			t.Fatal(err)
		}

		tests = append(tests, SugarTests{
			// # /ipns/<dnslink-fqdn>
			// test_localhost_gateway_response_should_contain \
			//   "request for localhost/ipns/{fqdn} redirects to DNSLink in subdomain" \
			//   "http://localhost:$GWAY_PORT/ipns/en.wikipedia-on-ipfs.org/wiki" \
			//   "Location: http://en.wikipedia-on-ipfs.org.ipns.localhost:$GWAY_PORT/wiki"
			{
				Name: "request for /ipns/{fqdn} redirects to DNSLink in subdomain",
				Request: Request().
					URL("{{url}}/ipns/{{fqdn}}/wiki/", gatewayURL, wikipedia),
				Response: Expect().
					Headers(
						Header("Location").
							Equals("{{scheme}}://{{fqdn}}.ipns.{{host}}/wiki/", u.Scheme, dnslink.InlineDNS(wikipedia), u.Host),
					),
			},
			// # <dnslink-fqdn>.ipns.localhost
			// # DNSLink test requires a daemon in online mode with precached /ipns/ mapping
			// test_kill_ipfs_daemon
			// DNSLINK_FQDN="dnslink-test.example.com"
			// export IPFS_NS_MAP="$DNSLINK_FQDN:/ipfs/$CIDv1"
			// test_launch_ipfs_daemon
			// test_localhost_gateway_response_should_contain \
			//   "request for {dnslink}.ipns.localhost returns expected payload" \
			//   "http://$DNSLINK_FQDN.ipns.localhost:$GWAY_PORT" \
			//   "$CID_VAL"
			{
				Name: "request for {dnslink}.ipns.{gateway} returns expected payload",
				Request: Request().
					URL("{{scheme}}://{{fqdn}}.ipns.{{host}}", u.Scheme, dnsLinkTest, u.Host),
				Response: Expect().
					Body("hello\n"),
			},
			// ## ============================================================================
			// ## Test DNSLink inlining on HTTP gateways
			// ## ============================================================================
			// # set explicit subdomain gateway config for the hostname
			// ipfs config --json Gateway.PublicGateways '{
			//   "localhost": {
			//     "UseSubdomains": true,
			//     "InlineDNSLink": true,
			//     "Paths": ["/ipfs", "/ipns", "/api"]
			//   },
			//   "example.com": {
			//     "UseSubdomains": true,
			//     "InlineDNSLink": true,
			//     "Paths": ["/ipfs", "/ipns", "/api"]
			//   }
			// }' || exit 1
			// # restart daemon to apply config changes
			// test_kill_ipfs_daemon
			// test_launch_ipfs_daemon_without_network

			// test_localhost_gateway_response_should_contain \
			//   "request for localhost/ipns/{fqdn} redirects to DNSLink in subdomain with DNS inlining" \
			//   "http://localhost:$GWAY_PORT/ipns/en.wikipedia-on-ipfs.org/wiki" \
			//   "Location: http://en-wikipedia--on--ipfs-org.ipns.localhost:$GWAY_PORT/wiki"

			// test_hostname_gateway_response_should_contain \
			//   "request for example.com/ipns/{fqdn} redirects to DNSLink in subdomain with DNS inlining" \
			//   "example.com" \
			//   "http://127.0.0.1:$GWAY_PORT/ipns/en.wikipedia-on-ipfs.org/wiki" \
			//   "Location: http://en-wikipedia--on--ipfs-org.ipns.example.com/wiki"

			// # example.com/ipns/<dnslink-fqdn>

			// test_hostname_gateway_response_should_contain \
			//   "request for example.com/ipns/{fqdn} redirects to DNSLink in subdomain" \
			//   "example.com" \
			//   "http://127.0.0.1:$GWAY_PORT/ipns/en.wikipedia-on-ipfs.org/wiki" \
			//   "Location: http://en.wikipedia-on-ipfs.org.ipns.example.com/wiki"

			// # DNSLink on Public gateway with a single-level wildcard TLS cert
			// # "Option C" from  https://github.com/ipfs/in-web-browsers/issues/169
			// test_expect_success \
			//   "request for example.com/ipns/{fqdn} with X-Forwarded-Proto redirects to TLS-safe label in subdomain" "
			//   curl -H \"Host: example.com\" -H \"X-Forwarded-Proto: https\" -sD - \"http://127.0.0.1:$GWAY_PORT/ipns/en.wikipedia-on-ipfs.org/wiki\" > response &&
			//   test_should_contain \"Location: https://en-wikipedia--on--ipfs-org.ipns.example.com/wiki\" response
			//   "
			{
				Name: "request for example.com/ipns/{fqdn} with X-Forwarded-Proto redirects to TLS-safe label in subdomain",
				Hint: `
				DNSLink on Public gateway with a single-level wildcard TLS cert
				"Option C" from https://github.com/ipfs/in-web-browsers/issues/169
				`,
				Request: Request().
					Header("X-Forwarded-Proto", "https").
					URL("{{url}}/ipns/{{wikipedia}}/wiki/", gatewayURL, wikipedia),
				Response: Expect().
					Headers(
						Header("Location").
							Equals("https://{{inlined}}.ipns.{{host}}/wiki/", dnslink.InlineDNS(wikipedia), u.Host),
					),
			},
			// # Support ipns:// in https://developer.mozilla.org/en-US/docs/Web/API/Navigator/registerProtocolHandler
			// test_hostname_gateway_response_should_contain \
			//   "request for example.com/ipns/?uri=ipns%3A%2F%2F.. produces redirect to /ipns/.. content path" \
			//   "example.com" \
			//   "http://127.0.0.1:$GWAY_PORT/ipns/?uri=ipns%3A%2F%2Fen.wikipedia-on-ipfs.org" \
			//   "Location: /ipns/en.wikipedia-on-ipfs.org"
			{
				Name: `request for example.com/ipns/?uri=ipns%3A%2F%2F.. produces redirect to /ipns/.. content path`,
				Hint: "Support ipns:// in https://developer.mozilla.org/en-US/docs/Web/API/Navigator/registerProtocolHandler",
				Request: Request().
					URL(`{{url}}/ipns/?uri=ipns%3A%2F%2F{{dnslink}}`, gatewayURL, wikipedia),
				Response: Expect().
					Headers(
						Header("Location").Equals("/ipns/{{wikipedia}}", wikipedia),
					),
			},
			// # DNSLink: <dnslink-fqdn>.ipns.example.com
			// # (not really useful outside of localhost, as setting TLS for more than one
			// # level of wildcard is a pain, but we support it if someone really wants it)
			// # ============================================================================

			// # DNSLink test requires a daemon in online mode with precached /ipns/ mapping
			// test_kill_ipfs_daemon
			// DNSLINK_FQDN="dnslink-subdomain-gw-test.example.org"
			// export IPFS_NS_MAP="$DNSLINK_FQDN:/ipfs/$CIDv1"
			// test_launch_ipfs_daemon

			// test_hostname_gateway_response_should_contain \
			//   "request for {dnslink}.ipns.example.com returns expected payload" \
			//   "$DNSLINK_FQDN.ipns.example.com" \
			//   "http://127.0.0.1:$GWAY_PORT" \
			//   "$CID_VAL"
			// Note: this test was merged with the test for wikipedia in the end.

			// # DNSLink on Public gateway with a single-level wildcard TLS cert
			// # "Option C" from  https://github.com/ipfs/in-web-browsers/issues/169
			// test_expect_success \
			//   "request for {single-label-dnslink}.ipns.example.com with X-Forwarded-Proto returns expected payload" "
			//   curl -H \"Host: dnslink--subdomain--gw--test-example-org.ipns.example.com\" -H \"X-Forwarded-Proto: https\" -sD - \"http://127.0.0.1:$GWAY_PORT\" > response &&
			//   test_should_contain \"$CID_VAL\" response
			//   "
			// Note: this test was merged with the test for wikipedia in the end.

		}...)
	}

	RunWithSpecs(t, helpers.UnwrapSubdomainTests(t, tests), specs.SubdomainGatewayIPNS)
}
