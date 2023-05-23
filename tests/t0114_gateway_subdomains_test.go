package tests

import (
	"net/url"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/helpers"
	"github.com/ipfs/gateway-conformance/tooling/ipns"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multicodec"
)

func TestGatewaySubdomains(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0114-gateway_subdomains.car")

	CIDVal := string(fixture.MustGetRawData("hello-CIDv1")) // hello
	DirCID := fixture.MustGetCid("testdirlisting")
	CIDv1 := fixture.MustGetCid("hello-CIDv1")
	CIDv0 := fixture.MustGetCid("hello-CIDv0")
	CIDv0to1 := fixture.MustGetCid("hello-CIDv0to1")
	CIDv1_TOO_LONG := fixture.MustGetCid("hello-CIDv1_TOO_LONG")
	CIDWikipedia := "QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco"

	tests := SugarTests{}

	// We're going to run the same test against multiple gateways (localhost, and a subdomain gateway)
	gatewayURLs := []string{
		SubdomainGatewayURL,
		SubdomainLocalhostGatewayURL,
	}

	for _, gatewayURL := range gatewayURLs {
		u, err := url.Parse(gatewayURL)
		if err != nil {
			t.Fatal(err)
		}

		tests = append(tests, SugarTests{
			{
				Name: "request for example.com/ipfs/{CIDv1} redirects to subdomain",
				Hint: `
					subdomains should not return payload directly,
					but redirect to URL with proper origin isolation
				`,
				Request: Request().DoNotFollowRedirects().URL("{{url}}/ipfs/{{cid}}/", gatewayURL, CIDv1),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").
							Hint("request for example.com/ipfs/{CIDv1} returns Location HTTP header for subdomain redirect in browsers").
							Contains("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, CIDv1, u.Host),
					).
					BodyWithHint(`
					We return body with HTTP 301 so existing cli scripts that use path-based
					gateway do not break (curl doesn't auto-redirect without passing -L; wget
					does not span across hostnames by default)
					Context: https://github.com/ipfs/go-ipfs/issues/6975
				`,
						IsEqual("hello\n"),
					),
			},
			{
				Name: "request for example.com/ipfs/{DirCID} redirects to subdomain",
				Hint: `
					subdomains should not return payload directly,
					but redirect to URL with proper origin isolation
				`,
				Request: Request().DoNotFollowRedirects().URL("{{url}}/ipfs/{{cid}}/", gatewayURL, DirCID),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").
							Hint("request for example.com/ipfs/{DirCID} returns Location HTTP header for subdomain redirect in browsers").
							Contains("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, DirCID, u.Host),
					),
			},
			{
				Name:    "request for example.com/ipfs/{CIDv0} redirects to CIDv1 representation in subdomain",
				Request: Request().DoNotFollowRedirects().URL("{{url}}/ipfs/{{cid}}/", gatewayURL, CIDv0),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").
							Hint("request for example.com/ipfs/{CIDv0to1} returns Location HTTP header for subdomain redirect in browsers").
							Contains("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, CIDv0to1, u.Host),
					),
			},
			// ============================================================================
			// Test subdomain-based requests to a local gateway with default config
			// (origin per content root at http://*.example.com)
			// ============================================================================
			{
				Name:    "request for {CID}.ipfs.example.com should return expected payload",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{cid}}.ipfs.{{host}}", u.Scheme, CIDv1, u.Host),
				Response: Expect().
					Status(200).
					Body(Contains(CIDVal)),
			},
			{
				Name:    "request for {CID}.ipfs.example.com/ipfs/{CID} should return HTTP 404",
				Hint:    "ensure /ipfs/ namespace is not mounted on subdomain",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{cid}}.ipfs.{{host}}/ipfs/{{cid}}", u.Scheme, CIDv1, u.Host),
				Response: Expect().
					Status(404),
			},
			{
				Name:    "request for {CID}.ipfs.example.com/ipfs/file.txt should return data from a file in CID content root",
				Hint:    "ensure requests to /ipfs/* are not blocked, if content root has such subdirectory",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{cid}}.ipfs.{{host}}/ipfs/file.txt", u.Scheme, DirCID, u.Host),
				Response: Expect().
					Status(200).
					Body(Contains("I am a txt file")),
			},
			{
				Name:    "valid file and subdirectory paths in directory listing at {cid}.ipfs.example.com",
				Hint:    "{CID}.ipfs.example.com/sub/dir (Directory Listing)",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, DirCID, u.Host),
				Response: Expect().
					Status(200).
					Body(And(
						// TODO: implement html expectations
						Contains(`<a href="/hello">hello</a>`),
						Contains(`<a href="/ipfs">ipfs</a>`),
					)),
			},
			{
				Name:    "valid parent directory path in directory listing at {cid}.ipfs.example.com/sub/dir",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{cid}}.ipfs.{{host}}/ipfs/ipns/", u.Scheme, DirCID, u.Host),
				Response: Expect().
					Status(200).
					Body(And(
						// TODO: implement html expectations
						Contains(`<a href="/ipfs/ipns/..">..</a>`),
						Contains(`<a href="/ipfs/ipns/bar">bar</a>`),
					)),
			},
			{
				Name:    "request for deep path resource at {cid}.ipfs.localhost/sub/dir/file",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{cid}}.ipfs.{{host}}/ipfs/ipns/bar", u.Scheme, DirCID, u.Host),
				Response: Expect().
					Status(200).
					Body(Contains("text-file-content")),
			},
			{
				Name: "valid breadcrumb links in the header of directory listing at {cid}.ipfs.example.com/sub/dir",
				Hint: `
			Note 1: we test for sneaky subdir names  {cid}.ipfs.example.com/ipfs/ipns/ :^)
			Note 2: example.com/ipfs/.. present in HTML will be redirected to subdomain, so this is expected behavior
			`,
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{cid}}.ipfs.{{host}}/ipfs/ipns/", u.Scheme, DirCID, u.Host),
				Response: Expect().
					Status(200).
					Body(
						And(
							Contains("Index of"),
							Contains(`/ipfs/<a href="//{{host}}/ipfs/{{cid}}">{{cid}}</a>/<a href="//{{host}}/ipfs/{{cid}}/ipfs">ipfs</a>/<a href="//{{host}}/ipfs/{{cid}}/ipfs/ipns">ipns</a>`,
								u.Host, DirCID),
						),
					),
			},
			// ## ============================================================================
			// ## Test subdomain-based requests with a custom hostname config
			// ## (origin per content root at http://*.example.com)
			// ## ============================================================================

			// # example.com/ip(f|n)s/*
			// # =============================================================================

			// # path requests to the root hostname should redirect
			// # to a subdomain URL with proper origin isolation

			{
				Name:    "request for example.com/ipfs/{CIDv1} produces redirect to {CIDv1}.ipfs.example.com",
				Hint:    "path requests to the root hostname should redirect to a subdomain URL with proper origin isolation",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{host}}/ipfs/{{cid}}/", u.Scheme, u.Host, CIDv1),
				Response: Expect().
					Headers(
						Header("Location").Equals("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, CIDv1, u.Host),
					),
			},

			{
				Name:    "request for example.com/ipfs/{InvalidCID} produces useful error before redirect",
				Hint:    "error message should include original CID (and it should be case-sensitive, as we can't assume everyone uses base32)",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{host}}/ipfs/QmInvalidCID", u.Scheme, u.Host),
				Response: Expect().
					Body(Contains(`invalid path "/ipfs/QmInvalidCID"`)),
			},

			{
				Name:    "request for example.com/ipfs/{CIDv0} produces redirect to {CIDv1}.ipfs.example.com",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{host}}/ipfs/{{cid}}/", u.Scheme, u.Host, CIDv0),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, CIDv0to1, u.Host),
					),
			},

			{
				Name: "request for http://example.com/ipfs/{CID} with X-Forwarded-Proto: https produces redirect to HTTPS URL",
				Hint: "Support X-Forwarded-Proto",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{host}}/ipfs/{{cid}}/", u.Scheme, u.Host, CIDv1).
					Header("X-Forwarded-Proto", "https"),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("https://{{cid}}.ipfs.{{host}}/", CIDv1, u.Host),
					),
			},

			{
				Name: "request for example.com/ipfs/?uri=ipfs%3A%2F%2F.. produces redirect to /ipfs/.. content path",
				Hint: "Support ipfs:// in https://developer.mozilla.org/en-US/docs/Web/API/Navigator/registerProtocolHandler",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{host}}/ipfs/", u.Scheme, u.Host).
					Query(
						"uri", "ipfs://{{host}}/wiki/Diego_Maradona.html", CIDWikipedia,
					),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("/ipfs/{{cid}}/wiki/Diego_Maradona.html", CIDWikipedia),
					),
			},
			{
				Name:    "request for a too long CID at localhost/ipfs/{CIDv1} returns human readable error",
				Hint:    "router should not redirect to hostnames that could fail due to DNS limits",
				Request: Request().DoNotFollowRedirects().URL("{{url}}/ipfs/{{cid}}", gatewayURL, CIDv1_TOO_LONG),
				Response: Expect().
					Status(400).
					Body(Contains("CID incompatible with DNS label length limit of 63")),
			},
			{
				Name:    "request for a too long CID at {CIDv1}.ipfs.localhost returns expected payload",
				Hint:    "direct request should also fail (provides the same UX as router and avoids confusion)",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, CIDv1_TOO_LONG, u.Host),
				Response: Expect().
					Status(400).
					Body(Contains("CID incompatible with DNS label length limit of 63")),
			},
			// ## ============================================================================
			// ## Test support for X-Forwarded-Host
			// ## ============================================================================
			{
				Name:    "request for http://fake.domain.com/ipfs/{CID} doesn't match the example.com gateway",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{domain}}/ipfs/{{cid}}", u.Scheme, "fake.domain.com", CIDv1),
				Response: Expect().
					Status(200),
			},
			{
				Name: "request for http://fake.domain.com/ipfs/{CID} with X-Forwarded-Host: example.com match the example.com gateway",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{domain}}/ipfs/{{cid}}", u.Scheme, "fake.domain.com", CIDv1).
					Header("X-Forwarded-Host", u.Host),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, CIDv1, u.Host),
					),
			},
			{
				Name: "request for http://fake.domain.com/ipfs/{CID} with X-Forwarded-Host: example.com and X-Forwarded-Proto: https match the example.com gateway, redirect with https",
				Request: Request().DoNotFollowRedirects().URL("{{scheme}}://{{domain}}/ipfs/{{cid}}", u.Scheme, "fake.domain.com", CIDv1).
					Header("X-Forwarded-Host", u.Host).
					Header("X-Forwarded-Proto", "https"),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("https://{{cid}}.ipfs.{{host}}/", CIDv1, u.Host),
					),
			},
		}...)
	}

	RunIfSpecsAreEnabled(t, helpers.UnwrapSubdomainTests(t, tests), specs.SubdomainGateway)
}

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
						DoNotFollowRedirects().
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
						DoNotFollowRedirects().
						URL("{{scheme}}://{{cid}}.ipns.{{host}}/", u.Scheme, record.IntoCID(multicodec.DagPb, multibase.Base36), u.Host),
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
				// Done above, thanks to the loop
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
					DoNotFollowRedirects().
					URL("{{url}}/ipns/{{cid}}", gatewayURL, ed25519Fixture.B58MH()),
				Response: Expect().
					Headers(
						Header("Location").
							Equals("{{scheme}}://{{cid}}.ipns.{{host}}/", u.Scheme, ed25519Fixture.IntoCID(multicodec.Libp2pKey, multibase.Base36), u.Host),
					),
			},
		}...)

	}

	RunIfSpecsAreEnabled(t, helpers.UnwrapSubdomainTests(t, tests), specs.SubdomainGateway, specs.IPNSResolver)
}

func TestGatewaySubdomainAndDnsLink(t *testing.T) {
	tests := SugarTests{}

	// We're going to run the same test against multiple gateways (localhost, and a subdomain gateway)
	gatewayURLs := []string{
		SubdomainGatewayURL,
		SubdomainLocalhostGatewayURL,
	}

	for _, gatewayURL := range gatewayURLs {
		_, err := url.Parse(gatewayURL)
		if err != nil {
			t.Fatal(err)
		}

		tests = append(tests, SugarTests{
			// # /ipns/<dnslink-fqdn>

			// test_localhost_gateway_response_should_contain \
			//   "request for localhost/ipns/{fqdn} redirects to DNSLink in subdomain" \
			//   "http://localhost:$GWAY_PORT/ipns/en.wikipedia-on-ipfs.org/wiki" \
			//   "Location: http://en.wikipedia-on-ipfs.org.ipns.localhost:$GWAY_PORT/wiki"

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

			// # Support ipns:// in https://developer.mozilla.org/en-US/docs/Web/API/Navigator/registerProtocolHandler
			// test_hostname_gateway_response_should_contain \
			//   "request for example.com/ipns/?uri=ipns%3A%2F%2F.. produces redirect to /ipns/.. content path" \
			//   "example.com" \
			//   "http://127.0.0.1:$GWAY_PORT/ipns/?uri=ipns%3A%2F%2Fen.wikipedia-on-ipfs.org" \
			//   "Location: /ipns/en.wikipedia-on-ipfs.org"

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

			// # DNSLink on Public gateway with a single-level wildcard TLS cert
			// # "Option C" from  https://github.com/ipfs/in-web-browsers/issues/169
			// test_expect_success \
			//   "request for {single-label-dnslink}.ipns.example.com with X-Forwarded-Proto returns expected payload" "
			//   curl -H \"Host: dnslink--subdomain--gw--test-example-org.ipns.example.com\" -H \"X-Forwarded-Proto: https\" -sD - \"http://127.0.0.1:$GWAY_PORT\" > response &&
			//   test_should_contain \"$CID_VAL\" response
			//   "

			// ## ============================================================================
			// ## Test DNSLink requests with a custom PublicGateway (hostname config)
			// ## (DNSLink site at http://dnslink-test.example.com)
			// ## ============================================================================
			// # disable wildcard DNSLink gateway
			// # and enable it on specific NSLink hostname
			// ipfs config --json Gateway.NoDNSLink true && \
			// ipfs config --json Gateway.PublicGateways '{
			//   "dnslink-enabled-on-fqdn.example.org": {
			//     "NoDNSLink": false,
			//     "UseSubdomains": false,
			//     "Paths": ["/ipfs"]
			//   },
			//   "only-dnslink-enabled-on-fqdn.example.org": {
			//     "NoDNSLink": false,
			//     "UseSubdomains": false,
			//     "Paths": []
			//   },
			//   "dnslink-disabled-on-fqdn.example.com": {
			//     "NoDNSLink": true,
			//     "UseSubdomains": false,
			//     "Paths": []
			//   }
			// }' || exit 1

			// # DNSLink test requires a daemon in online mode with precached /ipns/ mapping
			// DNSLINK_FQDN="dnslink-enabled-on-fqdn.example.org"
			// ONLY_DNSLINK_FQDN="only-dnslink-enabled-on-fqdn.example.org"
			// NO_DNSLINK_FQDN="dnslink-disabled-on-fqdn.example.com"
			// export IPFS_NS_MAP="$DNSLINK_FQDN:/ipfs/$CIDv1,$ONLY_DNSLINK_FQDN:/ipfs/$DIR_CID"

			// # DNSLink enabled

			// test_hostname_gateway_response_should_contain \
			//   "request for http://{dnslink-fqdn}/ PublicGateway returns expected payload" \
			//   "$DNSLINK_FQDN" \
			//   "http://127.0.0.1:$GWAY_PORT/" \
			//   "$CID_VAL"

			// test_hostname_gateway_response_should_contain \
			//   "request for {dnslink-fqdn}/ipfs/{cid} returns expected payload when /ipfs is on Paths whitelist" \
			//   "$DNSLINK_FQDN" \
			//   "http://127.0.0.1:$GWAY_PORT/ipfs/$CIDv1" \
			//   "$CID_VAL"

			// # Test for a fun edge case: DNSLink-only gateway without  /ipfs/ namespace
			// # mounted, and with subdirectory named "ipfs" ¯\_(ツ)_/¯
			// test_hostname_gateway_response_should_contain \
			//   "request for {dnslink-fqdn}/ipfs/file.txt returns data from content root when /ipfs in not on Paths whitelist" \
			//   "$ONLY_DNSLINK_FQDN" \
			//   "http://127.0.0.1:$GWAY_PORT/ipfs/file.txt" \
			//   "I am a txt file"

			// test_hostname_gateway_response_should_contain \
			//   "request for {dnslink-fqdn}/ipns/{peerid} returns 404 when path is not whitelisted" \
			//   "$DNSLINK_FQDN" \
			//   "http://127.0.0.1:$GWAY_PORT/ipns/$RSA_IPNS_IDv0" \
			//   "404 Not Found"

			// test_hostname_gateway_response_should_contain \
			//   "request for {dnslink-fqdn}/ipns/{peerid} returns 404 when path is not whitelisted" \
			//   "$DNSLINK_FQDN" \
			//   "http://127.0.0.1:$GWAY_PORT/ipns/$ED25519_IPNS_IDv0" \
			//   "404 Not Found"

			// # DNSLink disabled

			// test_hostname_gateway_response_should_contain \
			//   "request for http://{dnslink-fqdn}/ returns 404 when NoDNSLink=true" \
			//   "$NO_DNSLINK_FQDN" \
			//   "http://127.0.0.1:$GWAY_PORT/" \
			//   "404 Not Found"

			// test_hostname_gateway_response_should_contain \
			//   "request for {dnslink-fqdn}/ipfs/{cid} returns 404 when path is not whitelisted" \
			//   "$NO_DNSLINK_FQDN" \
			//   "http://127.0.0.1:$GWAY_PORT/ipfs/$CIDv0" \
			//   "404 Not Found"

			// ## ============================================================================
			// ## Test wildcard DNSLink (any hostname, with default config)
			// ## ============================================================================

			// test_kill_ipfs_daemon

			// # enable wildcard DNSLink gateway (any value in Host header)
			// # and remove custom PublicGateways
			// ipfs config --json Gateway.NoDNSLink false && \
			// ipfs config --json Gateway.PublicGateways '{}' || exit 1

			// # DNSLink test requires a daemon in online mode with precached /ipns/ mapping
			// DNSLINK_FQDN="wildcard-dnslink-not-in-config.example.com"
			// export IPFS_NS_MAP="$DNSLINK_FQDN:/ipfs/$CIDv1"

			// # restart daemon to apply config changes
			// test_launch_ipfs_daemon

			// # make sure test setup is valid (fail if CoreAPI is unable to resolve)
			// test_expect_success "spoofed DNSLink record resolves in cli" "
			//   ipfs resolve /ipns/$DNSLINK_FQDN > result &&
			//   test_should_contain \"$CIDv1\" result &&
			//   ipfs cat /ipns/$DNSLINK_FQDN > result &&
			//   test_should_contain \"$CID_VAL\" result
			// "

			// # gateway test
			//
			//	test_hostname_gateway_response_should_contain \
			//	  "request for http://{dnslink-fqdn}/ (wildcard) returns expected payload" \
			//	  "$DNSLINK_FQDN" \
			//	  "http://127.0.0.1:$GWAY_PORT/" \
			//	  "$CID_VAL"
		}...)
	}

	RunIfSpecsAreEnabled(t, helpers.UnwrapSubdomainTests(t, tests), specs.SubdomainGateway, specs.DNSLinkResolver)
}
