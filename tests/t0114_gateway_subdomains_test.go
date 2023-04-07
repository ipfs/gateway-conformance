package tests

import (
	"net/url"
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/dnslink"
	"github.com/ipfs/gateway-conformance/tooling/helpers"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	"github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func fqdnEncoding(domain string) string {
	// TODO: verify the encoding
	// https://github.com/ipfs/in-web-browsers/issues/169
	x := strings.ReplaceAll(domain, "-", "--")
	x = strings.ReplaceAll(x, ".", "-")
	return x
}

func TestGatewaySubdomains(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0114-gateway_subdomains.car")

	CIDVal := string(fixture.MustGetRawData("hello-CIDv1")) // hello
	DirCID := fixture.MustGetCid("testdirlisting")
	CIDv1 := fixture.MustGetCid("hello-CIDv1")
	CIDv0 := fixture.MustGetCid("hello-CIDv0")
	CIDv0to1 := fixture.MustGetCid("hello-CIDv0to1")
	CIDv1_TOO_LONG := fixture.MustGetCid("hello-CIDv1_TOO_LONG")
	CIDWikipedia := "QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco"

	dnsLinks := dnslink.MustOpenDNSLink("t0114-gateway_subdomains.yml")
	wikipediaDnsLink := dnsLinks.Get("wikipedia")
	fqdnDnsLink := dnsLinks.Get("fqdn")

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

		fqdn := fqdnDnsLink
		wikipedia := wikipediaDnsLink
		wikipediaEncoded := fqdnEncoding(wikipedia)

		// TODO(lidel): let's chat how we want to implement this.
		// What I understand: we have an encoding to make sure an ipns/some-domain is a single domain (a.b.c -> a-b-c)
		// localhost has a special behavior, every subdomain redirects is a single domain.
		wikipediaSubdomain := wikipediaDnsLink

		if gatewayURL == SubdomainLocalhostGatewayURL {
			wikipediaSubdomain = wikipediaEncoded
		}

		tests = append(tests, test.SugarTests{
			// with(testGatewayWithManyProtocols(t,
			// 	"request for example.com/ipfs/{CIDv1} redirects to subdomain",
			// 	`
			// 	subdomains should not return payload directly,
			// 	but redirect to URL with proper origin isolation
			// 	`,
			// 	URL("%s/ipfs/%s/", gatewayURL, CIDv1),
			// 	Expect().
			// 		Status(301).
			// 		Headers(
			// 			Header("Location").
			// 				Hint("request for example.com/ipfs/{CIDv1} returns Location HTTP header for subdomain redirect in browsers").
			// 				Contains("%s://%s.ipfs.%s/", u.Scheme, CIDv1, u.Host),
			// 		).
			// 		BodyWithHint(`
			// 			We return body with HTTP 301 so existing cli scripts that use path-based
			// 			gateway do not break (curl doesn't auto-redirect without passing -L; wget
			// 			does not span across hostnames by default)
			// 			Context: https://github.com/ipfs/go-ipfs/issues/6975
			// 		`,
			// 			IsEqual("hello\n"),
			// 		),
			// ))
			{
				Name: "request for example.com/ipfs/{CIDv1} redirects to subdomain",
				Hint: `
				subdomains should not return payload directly,
				but redirect to URL with proper origin isolation
				`,
				Request: Request().DoNotFollowRedirects().
					URL("%s/ipfs/%s/", gatewayURL, CIDv1),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").
							Hint("request for example.com/ipfs/{CIDv1} returns Location HTTP header for subdomain redirect in browsers").
							Contains("%s://%s.ipfs.%s/", u.Scheme, CIDv1, u.Host),
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
			// with(testGatewayWithManyProtocols(t,
			// 	"request for example.com/ipfs/{DirCID} redirects to subdomain",
			// 	`
			// 	subdomains should not return payload directly,
			// 	but redirect to URL with proper origin isolation
			// 	`,
			// 	URL("%s/ipfs/%s/", gatewayURL, DirCID),
			// 	Expect().
			// 		Status(301).
			// 		Headers(
			// 			Header("Location").
			// 				Hint("request for example.com/ipfs/{DirCID} returns Location HTTP header for subdomain redirect in browsers").
			// 				Contains("%s://%s.ipfs.%s/", u.Scheme, DirCID, u.Host),
			// 		),
			// ))
			{
				Name: "request for example.com/ipfs/{DirCID} redirects to subdomain",
				Hint: `
				subdomains should not return payload directly,
				but redirect to URL with proper origin isolation
				`,
				Request: Request().DoNotFollowRedirects().
					URL("%s/ipfs/%s/", gatewayURL, DirCID),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").
							Hint("request for example.com/ipfs/{DirCID} returns Location HTTP header for subdomain redirect in browsers").
							Contains("%s://%s.ipfs.%s/", u.Scheme, DirCID, u.Host),
					),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"request for example.com/ipfs/{CIDv0} redirects to CIDv1 representation in subdomain",
			// 	"",
			// 	URL("%s/ipfs/%s/", gatewayURL, CIDv0),
			// 	Expect().
			// 		Status(301).
			// 		Headers(
			// 			Header("Location").
			// 				Hint("request for example.com/ipfs/{CIDv0to1} returns Location HTTP header for subdomain redirect in browsers").
			// 				Contains("%s://%s.ipfs.%s/", u.Scheme, CIDv0to1, u.Host),
			// 		),
			// ))
			{
				Name: "request for example.com/ipfs/{CIDv0} redirects to CIDv1 representation in subdomain",
				Request: Request().DoNotFollowRedirects().
					URL("%s/ipfs/%s/", gatewayURL, CIDv0),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").
							Hint("request for example.com/ipfs/{CIDv0to1} returns Location HTTP header for subdomain redirect in browsers").
							Contains("%s://%s.ipfs.%s/", u.Scheme, CIDv0to1, u.Host),
					),
			},

			// # /ipns/<libp2p-key>

			// test_localhost_gateway_response_should_contain \
			//   "request for localhost/ipns/{CIDv0} redirects to CIDv1 with libp2p-key multicodec in subdomain" \
			//   "http://localhost:$GWAY_PORT/ipns/$RSA_IPNS_IDv0" \
			//   "Location: http://${RSA_IPNS_IDv1}.ipns.localhost:$GWAY_PORT/"
			// TODO: ipns
			// test_localhost_gateway_response_should_contain \
			//   "request for localhost/ipns/{CIDv0} redirects to CIDv1 with libp2p-key multicodec in subdomain" \
			//   "http://localhost:$GWAY_PORT/ipns/$ED25519_IPNS_IDv0" \
			//   "Location: http://${ED25519_IPNS_IDv1}.ipns.localhost:$GWAY_PORT/"
			// TODO: ipns

			// # /ipns/<dnslink-fqdn>
			// test_localhost_gateway_response_should_contain \
			//   "request for localhost/ipns/{fqdn} redirects to DNSLink in subdomain" \
			//   "http://localhost:$GWAY_PORT/ipns/en.wikipedia-on-ipfs.org/wiki" \
			//   "Location: http://en.wikipedia-on-ipfs.org.ipns.localhost:$GWAY_PORT/wiki"
			{
				Name: "request for localhost/ipns/{fqdn} redirects to DNSLink in subdomain",
				Request: Request().DoNotFollowRedirects().
					// TOOD(lidel): I have to use trailing / here, else
					// I receive a redirect to /ipns/%s/wiki/
					URL("%s/ipns/%s/wiki/", gatewayURL, wikipedia),
				Response: Expect().
					Status(301).
					Headers(
						// TODO(lidel): We have to use a different encoding
						// for the subdomain, I'm not sure about how maintainable
						// this is.
						Header("Location").
							Contains("%s://%s.ipns.%s/wiki/", u.Scheme, wikipediaSubdomain, u.Host),
					),
			},
			// // ============================================================================
			// // Test subdomain-based requests to a local gateway with default config
			// // (origin per content root at http://*.example.com)
			// // ============================================================================

			// with(testGatewayWithManyProtocols(t,
			// 	"request for {CID}.ipfs.example.com should return expected payload",
			// 	"",
			// 	URL("%s://%s.ipfs.%s", u.Scheme, CIDv1, u.Host),
			// 	Expect().
			// 		Status(200).
			// 		Body(Contains(CIDVal)),
			// ))
			{
				Name: "request for {CID}.ipfs.example.com should return expected payload",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s.ipfs.%s", u.Scheme, CIDv1, u.Host),
				Response: Expect().
					Status(200).
					Body(Contains(CIDVal)),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"request for {CID}.ipfs.example.com/ipfs/{CID} should return HTTP 404",
			// 	"ensure /ipfs/ namespace is not mounted on subdomain",
			// 	URL("%s://%s.ipfs.%s/ipfs/%s", u.Scheme, CIDv1, u.Host, CIDv1),
			// 	Expect().
			// 		Status(404),
			// ))
			{
				Name: "request for {CID}.ipfs.example.com/ipfs/{CID} should return HTTP 404",
				Hint: "ensure /ipfs/ namespace is not mounted on subdomain",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s.ipfs.%s/ipfs/%s", u.Scheme, CIDv1, u.Host, CIDv1),
				Response: Expect().
					Status(404),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"request for {CID}.ipfs.example.com/ipfs/file.txt should return data from a file in CID content root",
			// 	"ensure requests to /ipfs/* are not blocked, if content root has such subdirectory",
			// 	URL("%s://%s.ipfs.%s/ipfs/file.txt", u.Scheme, DirCID, u.Host),
			// 	Expect().
			// 		Status(200).
			// 		Body(Contains("I am a txt file")),
			// ))
			{
				Name: "request for {CID}.ipfs.example.com/ipfs/file.txt should return data from a file in CID content root",
				Hint: "ensure requests to /ipfs/* are not blocked, if content root has such subdirectory",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s.ipfs.%s/ipfs/file.txt", u.Scheme, DirCID, u.Host),
				Response: Expect().
					Status(200).
					Body(Contains("I am a txt file")),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"valid file and subdirectory paths in directory listing at {cid}.ipfs.example.com",
			// 	"{CID}.ipfs.example.com/sub/dir (Directory Listing)",
			// 	URL("%s://%s.ipfs.%s/", u.Scheme, DirCID, u.Host),
			// 	Expect().
			// 		Status(200).
			// 		Body(And(
			// 			// TODO: implement html expectations
			// 			Contains("<a href=\"/hello\">hello</a>"),
			// 			Contains("<a href=\"/ipfs\">ipfs</a>"),
			// 		)),
			// ))
			{
				Name: "valid file and subdirectory paths in directory listing at {cid}.ipfs.example.com",
				Hint: "{CID}.ipfs.example.com/sub/dir (Directory Listing)",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s.ipfs.%s/", u.Scheme, DirCID, u.Host),
				Response: Expect().
					Status(200).
					Body(And(
						// TODO: implement html expectations
						Contains("<a href=\"/hello\">hello</a>"),
						Contains("<a href=\"/ipfs\">ipfs</a>"),
					)),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"valid parent directory path in directory listing at {cid}.ipfs.example.com/sub/dir",
			// 	"",
			// 	URL("%s://%s.ipfs.%s/ipfs/ipns/", u.Scheme, DirCID, u.Host),
			// 	Expect().
			// 		Status(200).
			// 		Body(And(
			// 			// TODO: implement html expectations
			// 			Contains("<a href=\"/ipfs/ipns/..\">..</a>"),
			// 			Contains("<a href=\"/ipfs/ipns/bar\">bar</a>"),
			// 		)),
			// ))
			{
				Name: "valid parent directory path in directory listing at {cid}.ipfs.example.com/sub/dir",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s.ipfs.%s/ipfs/ipns/", u.Scheme, DirCID, u.Host),
				Response: Expect().
					Status(200).
					Body(And(
						// TODO: implement html expectations
						Contains("<a href=\"/ipfs/ipns/..\">..</a>"),
						Contains("<a href=\"/ipfs/ipns/bar\">bar</a>"),
					)),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"request for deep path resource at {cid}.ipfs.localhost/sub/dir/file",
			// 	"",
			// 	URL("%s://%s.ipfs.%s/ipfs/ipns/bar", u.Scheme, DirCID, u.Host),
			// 	Expect().
			// 		Status(200).
			// 		Body(Contains("text-file-content")),
			// ))
			{
				Name: "request for deep path resource at {cid}.ipfs.localhost/sub/dir/file",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s.ipfs.%s/ipfs/ipns/bar", u.Scheme, DirCID, u.Host),
				Response: Expect().
					Status(200).
					Body(Contains("text-file-content")),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"valid breadcrumb links in the header of directory listing at {cid}.ipfs.example.com/sub/dir",
			// 	`
			// 	Note 1: we test for sneaky subdir names  {cid}.ipfs.example.com/ipfs/ipns/ :^)
			// 	Note 2: example.com/ipfs/.. present in HTML will be redirected to subdomain, so this is expected behavior
			// 	`,
			// 	URL("%s://%s.ipfs.%s/ipfs/ipns/", u.Scheme, DirCID, u.Host),
			// 	Expect().
			// 		Status(200).
			// 		Body(
			// 			And(
			// 				Contains("Index of"),
			// 				Contains("/ipfs/<a href=\"//%s/ipfs/%s\">%s</a>/<a href=\"//%s/ipfs/%s/ipfs\">ipfs</a>/<a href=\"//%s/ipfs/%s/ipfs/ipns\">ipns</a>",
			// 					u.Host, DirCID, DirCID, u.Host, DirCID, u.Host, DirCID),
			// 			),
			// 		),
			// ))
			{
				Name: "valid breadcrumb links in the header of directory listing at {cid}.ipfs.example.com/sub/dir",
				Hint: `
				Note 1: we test for sneaky subdir names  {cid}.ipfs.example.com/ipfs/ipns/ :^)
				Note 2: example.com/ipfs/.. present in HTML will be redirected to subdomain, so this is expected behavior
				`,
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s.ipfs.%s/ipfs/ipns/", u.Scheme, DirCID, u.Host),
				Response: Expect().
					Status(200).
					Body(
						And(
							Contains("Index of"),
							Contains("/ipfs/<a href=\"//%s/ipfs/%s\">%s</a>/<a href=\"//%s/ipfs/%s/ipfs\">ipfs</a>/<a href=\"//%s/ipfs/%s/ipfs/ipns\">ipns</a>",
								u.Host, DirCID, DirCID, u.Host, DirCID, u.Host, DirCID),
						),
					),
			},

			// // TODO: # *.ipns.localhost
			// // TODO: # <libp2p-key>.ipns.localhost
			// // TODO: # <dnslink-fqdn>.ipns.localhost

			// # DNSLink test requires a daemon in online mode with precached /ipns/ mapping
			// test_kill_ipfs_daemon
			// DNSLINK_FQDN="dnslink-test.example.com"
			// export IPFS_NS_MAP="$DNSLINK_FQDN:/ipfs/$CIDv1"
			// test_launch_ipfs_daemon

			//	test_localhost_gateway_response_should_contain \
			//	  "request for {dnslink}.ipns.localhost returns expected payload" \
			//	  "http://$DNSLINK_FQDN.ipns.localhost:$GWAY_PORT" \
			//	  "$CID_VAL"
			{
				Name: "request for {dnslink}.ipns.localhost returns expected payload",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s.ipns.%s", u.Scheme, fqdn, u.Host),
				Response: Expect().
					Status(200).
					Body(Contains(CIDVal)),
			},
			// // ## ============================================================================
			// // ## Test DNSLink inlining on HTTP gateways
			// // ## ============================================================================

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

			//	test_localhost_gateway_response_should_contain \
			//	  "request for localhost/ipns/{fqdn} redirects to DNSLink in subdomain with DNS inlining" \
			//	  "http://localhost:$GWAY_PORT/ipns/en.wikipedia-on-ipfs.org/wiki" \
			//	  "Location: http://en-wikipedia--on--ipfs-org.ipns.localhost:$GWAY_PORT/wiki"
			{
				Name: "request for localhost/ipns/{fqdn} redirects to DNSLink in subdomain with DNS inlining",
				Request: Request().DoNotFollowRedirects().
					URL("%s/ipns/%s/wiki/", gatewayURL, wikipedia),
				Response: Expect().
					Status(301).
					Headers(
						// TODO(lidel): agree on a better name / approach for the wikipediaSubdomain which may be safe or not safe.
						Header("Location").Equals("%s://%s.ipns.%s/wiki/", u.Scheme, wikipediaSubdomain, u.Host),
					),
			},
			// test_hostname_gateway_response_should_contain \
			//   "request for example.com/ipns/{fqdn} redirects to DNSLink in subdomain with DNS inlining" \
			//   "example.com" \
			//   "http://127.0.0.1:$GWAY_PORT/ipns/en.wikipedia-on-ipfs.org/wiki" \
			//   "Location: http://en-wikipedia--on--ipfs-org.ipns.example.com/wiki"
			// TODO: this is hostname gateway

			// // ## ============================================================================
			// // ## Test subdomain-based requests with a custom hostname config
			// // ## (origin per content root at http://*.example.com)
			// // ## ============================================================================

			// // # example.com/ip(f|n)s/*
			// // # =============================================================================

			// // # path requests to the root hostname should redirect
			// // # to a subdomain URL with proper origin isolation

			// with(testGatewayWithManyProtocols(t,
			// 	"request for example.com/ipfs/{CIDv1} produces redirect to {CIDv1}.ipfs.example.com",
			// 	"path requests to the root hostname should redirect to a subdomain URL with proper origin isolation",
			// 	URL("%s://%s/ipfs/%s/", u.Scheme, u.Host, CIDv1),
			// 	Expect().
			// 		Headers(
			// 			Header("Location").Equals("%s://%s.ipfs.%s/", u.Scheme, CIDv1, u.Host),
			// 		),
			// ))
			{
				Name: "request for example.com/ipfs/{CIDv1} produces redirect to {CIDv1}.ipfs.example.com",
				Hint: "path requests to the root hostname should redirect to a subdomain URL with proper origin isolation",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s/ipfs/%s/", u.Scheme, u.Host, CIDv1),
				Response: Expect().
					Headers(
						Header("Location").Equals("%s://%s.ipfs.%s/", u.Scheme, CIDv1, u.Host),
					),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"request for example.com/ipfs/{InvalidCID} produces useful error before redirect",
			// 	"error message should include original CID (and it should be case-sensitive, as we can't assume everyone uses base32)",
			// 	URL("%s://%s/ipfs/QmInvalidCID", u.Scheme, u.Host),
			// 	Expect().
			// 		Body(Contains("invalid path \"/ipfs/QmInvalidCID\"")),
			// ))
			{
				Name: "request for example.com/ipfs/{InvalidCID} produces useful error before redirect",
				Hint: "error message should include original CID (and it should be case-sensitive, as we can't assume everyone uses base32)",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s/ipfs/QmInvalidCID", u.Scheme, u.Host),
				Response: Expect().
					Body(Contains("invalid path \"/ipfs/QmInvalidCID\"")),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"request for example.com/ipfs/{CIDv0} produces redirect to {CIDv1}.ipfs.example.com",
			// 	"",
			// 	URL("%s://%s/ipfs/%s/", u.Scheme, u.Host, CIDv0),
			// 	Expect().
			// 		Status(301).
			// 		Headers(
			// 			Header("Location").Equals("%s://%s.ipfs.%s/", u.Scheme, CIDv0to1, u.Host),
			// 		),
			// ))
			{
				Name: "request for example.com/ipfs/{CIDv0} produces redirect to {CIDv1}.ipfs.example.com",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s/ipfs/%s/", u.Scheme, u.Host, CIDv0),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("%s://%s.ipfs.%s/", u.Scheme, CIDv0to1, u.Host),
					),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"request for http://example.com/ipfs/{CID} with X-Forwarded-Proto: https produces redirect to HTTPS URL",
			// 	"Support X-Forwarded-Proto",
			// 	Request().
			// 		URL("%s://%s/ipfs/%s/", u.Scheme, u.Host, CIDv1).
			// 		Header("X-Forwarded-Proto", "https"),
			// 	Expect().
			// 		Status(301).
			// 		Headers(
			// 			Header("Location").Equals("https://%s.ipfs.%s/", CIDv1, u.Host),
			// 		),
			// ))
			{
				Name: "request for http://example.com/ipfs/{CID} with X-Forwarded-Proto: https produces redirect to HTTPS URL",
				Hint: "Support X-Forwarded-Proto",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s/ipfs/%s/", u.Scheme, u.Host, CIDv1).
					Header("X-Forwarded-Proto", "https"),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("https://%s.ipfs.%s/", CIDv1, u.Host),
					),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"request for example.com/ipfs/?uri=ipfs%3A%2F%2F.. produces redirect to /ipfs/.. content path",
			// 	"Support ipfs:// in https://developer.mozilla.org/en-US/docs/Web/API/Navigator/registerProtocolHandler",
			// 	Request().
			// 		URL("%s://%s/ipfs/", u.Scheme, u.Host).
			// 		Query(
			// 			"uri", "ipfs://%s/wiki/Diego_Maradona.html", CIDWikipedia,
			// 		),
			// 	Expect().
			// 		Status(301).
			// 		Headers(
			// 			Header("Location").Equals("/ipfs/%s/wiki/Diego_Maradona.html", CIDWikipedia),
			// 		),
			// ))
			{
				Name: "request for example.com/ipfs/?uri=ipfs%3A%2F%2F.. produces redirect to /ipfs/.. content path",
				Hint: "Support ipfs:// in https://developer.mozilla.org/en-US/docs/Web/API/Navigator/registerProtocolHandler",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s/ipfs/", u.Scheme, u.Host).
					Query(
						"uri", "ipfs://%s/wiki/Diego_Maradona.html", CIDWikipedia,
					),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("/ipfs/%s/wiki/Diego_Maradona.html", CIDWikipedia),
					),
			},
			// // # example.com/ipns/<libp2p-key>
			// // TODO

			// // # example.com/ipns/<dnslink-fqdn>
			// test_hostname_gateway_response_should_contain \
			//   "request for example.com/ipns/{fqdn} redirects to DNSLink in subdomain" \
			//   "example.com" \
			//   "http://127.0.0.1:$GWAY_PORT/ipns/en.wikipedia-on-ipfs.org/wiki" \
			//   "Location: http://en.wikipedia-on-ipfs.org.ipns.example.com/wiki"
			// SKIPPED: covered above

			// // # DNSLink on Public gateway with a single-level wildcard TLS cert
			// // # "Option C" from  https://github.com/ipfs/in-web-browsers/issues/169
			// test_expect_success \
			//   "request for example.com/ipns/{fqdn} with X-Forwarded-Proto redirects to TLS-safe label in subdomain" "
			//   curl -H \"Host: example.com\" -H \"X-Forwarded-Proto: https\" -sD - \"http://127.0.0.1:$GWAY_PORT/ipns/en.wikipedia-on-ipfs.org/wiki\" > response &&
			//   test_should_contain \"Location: https://en-wikipedia--on--ipfs-org.ipns.example.com/wiki\" response
			//   "
			{
				Name: "request for example.com/ipns/{fqdn} with X-Forwarded-Proto redirects to TLS-safe label in subdomain",
				Hint: `
				DNSLink on Public gateway with a single-level wildcard TLS cert
				"Option C" from  https://github.com/ipfs/in-web-browsers/issues/169
				`,
				Request: Request().DoNotFollowRedirects().
					URL("%s/ipns/%s/wiki/", gatewayURL, wikipedia).
					Header("X-Forwarded-Proto", "https"),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("https://%s.ipns.%s/wiki/", wikipediaEncoded, u.Host),
					),
			},
			// // # Support ipns:// in https://developer.mozilla.org/en-US/docs/Web/API/Navigator/registerProtocolHandler
			// // TODO

			// // # *.ipns.example.com
			// // # ============================================================================

			// // # <libp2p-key>.ipns.example.com

			// // # API on subdomain gateway example.com
			// // # ============================================================================

			// // # DNSLink: <dnslink-fqdn>.ipns.example.com
			// // # (not really useful outside of localhost, as setting TLS for more than one
			// // # level of wildcard is a pain, but we support it if someone really wants it)
			// // # ============================================================================

			// // # DNSLink on Public gateway with a single-level wildcard TLS cert
			// // # "Option C" from  https://github.com/ipfs/in-web-browsers/issues/169
			// // TODO(lidel): double check but should be covered previously

			// // ## Test subdomain handling of CIDs that do not fit in a single DNS Label (>63chars)
			// // ## https://github.com/ipfs/go-ipfs/issues/7318
			// // ## ============================================================================
			// // TODO

			// with(testGatewayWithManyProtocols(t,
			// 	"request for a too long CID at localhost/ipfs/{CIDv1} returns human readable error",
			// 	"router should not redirect to hostnames that could fail due to DNS limits",
			// 	URL("%s/ipfs/%s", gatewayURL, CIDv1_TOO_LONG),
			// 	Expect().
			// 		Status(400).
			// 		Body(Contains("CID incompatible with DNS label length limit of 63")),
			// ))
			{
				Name: "request for a too long CID at localhost/ipfs/{CIDv1} returns human readable error",
				Hint: "router should not redirect to hostnames that could fail due to DNS limits",
				Request: Request().DoNotFollowRedirects().
					URL("%s/ipfs/%s", gatewayURL, CIDv1_TOO_LONG),
				Response: Expect().
					Status(400).
					Body(Contains("CID incompatible with DNS label length limit of 63")),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"request for a too long CID at {CIDv1}.ipfs.localhost returns expected payload",
			// 	"direct request should also fail (provides the same UX as router and avoids confusion)",
			// 	URL("%s://%s.ipfs.%s/", u.Scheme, CIDv1_TOO_LONG, u.Host),
			// 	Expect().
			// 		Status(400).
			// 		Body(Contains("CID incompatible with DNS label length limit of 63")),
			// ))
			{
				Name: "request for a too long CID at {CIDv1}.ipfs.localhost returns expected payload",
				Hint: "direct request should also fail (provides the same UX as router and avoids confusion)",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s.ipfs.%s/", u.Scheme, CIDv1_TOO_LONG, u.Host),
				Response: Expect().
					Status(400).
					Body(Contains("CID incompatible with DNS label length limit of 63")),
			},
			// // # public subdomain gateway: *.example.com
			// // TODO: IPNS

			// // # Disable selected Paths for the subdomain gateway hostname
			// // # =============================================================================

			// // # disable /ipns for the hostname by not whitelisting it

			// // # refuse requests to Paths that were not explicitly whitelisted for the hostname

			// // MANY TODOs here

			// // ## ============================================================================
			// // ## Test support for X-Forwarded-Host
			// // ## ============================================================================

			// with(testGatewayWithManyProtocols(t,
			// 	"request for http://fake.domain.com/ipfs/{CID} doesn't match the example.com gateway",
			// 	"",
			// 	URL("%s://%s/ipfs/%s", u.Scheme, "fake.domain.com", CIDv1),
			// 	Expect().
			// 		Status(200),
			// ))
			{
				Name: "request for http://fake.domain.com/ipfs/{CID} doesn't match the example.com gateway",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s/ipfs/%s", u.Scheme, "fake.domain.com", CIDv1),
				Response: Expect().
					Status(200),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"request for http://fake.domain.com/ipfs/{CID} with X-Forwarded-Host: example.com match the example.com gateway",
			// 	"",
			// 	Request().
			// 		URL("%s://%s/ipfs/%s", u.Scheme, "fake.domain.com", CIDv1).
			// 		Header("X-Forwarded-Host", u.Host),
			// 	Expect().
			// 		Status(301).
			// 		Headers(
			// 			Header("Location").Equals("%s://%s.ipfs.%s/", u.Scheme, CIDv1, u.Host),
			// 		),
			// ))
			{
				Name: "request for http://fake.domain.com/ipfs/{CID} with X-Forwarded-Host: example.com match the example.com gateway",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s/ipfs/%s", u.Scheme, "fake.domain.com", CIDv1).
					Header("X-Forwarded-Host", u.Host),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("%s://%s.ipfs.%s/", u.Scheme, CIDv1, u.Host),
					),
			},
			// with(testGatewayWithManyProtocols(t,
			// 	"request for http://fake.domain.com/ipfs/{CID} with X-Forwarded-Host: example.com and X-Forwarded-Proto: https match the example.com gateway, redirect with https",
			// 	"",
			// 	Request().
			// 		URL("%s://%s/ipfs/%s", u.Scheme, "fake.domain.com", CIDv1).
			// 		Header("X-Forwarded-Host", u.Host).
			// 		Header("X-Forwarded-Proto", "https"),
			// 	Expect().
			// 		Status(301).
			// 		Headers(
			// 			Header("Location").Equals("https://%s.ipfs.%s/", CIDv1, u.Host),
			// 		),
			// ))
			{
				Name: "request for http://fake.domain.com/ipfs/{CID} with X-Forwarded-Host: example.com and X-Forwarded-Proto: https match the example.com gateway, redirect with https",
				Request: Request().DoNotFollowRedirects().
					URL("%s://%s/ipfs/%s", u.Scheme, "fake.domain.com", CIDv1).
					Header("X-Forwarded-Host", u.Host).
					Header("X-Forwarded-Proto", "https"),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("https://%s.ipfs.%s/", CIDv1, u.Host),
					),
			},
		}...)
	}

	if specs.SubdomainGateway.IsEnabled() {
		Run(t, helpers.UnwrapSubdomainTests(t, tests))
	} else {
		t.Skip("subdomain gateway disabled")
	}
}
