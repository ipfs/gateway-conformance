package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/dnslink"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

// TestDNSLinkGatewayIpfsUriAuthority asserts the Ipfs-Uri authority rules
// for DNSLink hosts (IPIP-548): a dotted name becomes an ipns:// URI, and
// a DNSLink name with no dot never does, because on each network such a
// name can point at different content.
func TestDNSLinkGatewayIpfsUriAuthority(t *testing.T) {
	tooling.LogTestGroup(t, GroupDNSLink)

	dnsLinks := dnslink.MustOpenDNSLink("dir_listing/dnslink.yml")
	dotted := dnsLinks.MustGet("ipfs-uri-dotted-name")
	dotless := dnsLinks.MustGet("ipfs-uri-name-with-no-dot")

	tests := SugarTests{
		{
			Name: "GET with a dotted DNSLink Host returns Ipfs-Uri with the name as authority",
			Hint: "A DNSLink name with at least one dot, including private network names, is a valid ipns:// authority",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#ipfs-uri-response-header",
			Request: Request().
				Path("/").
				Header("Host", dotted),
			Response: Expect().
				Status(200).
				Headers(
					Header("Ipfs-Uri").
						Equals("ipns://{{host}}/", dotted),
				),
		},
		{
			Name: "GET with a DNSLink Host that has no dot returns no Ipfs-Uri header",
			Hint: "A DNSLink name with no dot can point at different content on each network, so it is never an ipns:// authority; the gateway omits the header no matter how it answers the request",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#ipfs-uri-response-header",
			Request: Request().
				Path("/").
				Header("Host", dotless),
			Response: Expect().
				Headers(
					Header("Ipfs-Uri").
						IsEmpty(),
				),
		},
	}

	RunWithSpecs(t, tests, specs.DNSLinkGateway)
}

func TestDNSLinkGatewayUnixFSDirectoryListing(t *testing.T) {
	tooling.LogTestGroup(t, GroupDNSLink)

	fixture := car.MustOpenUnixfsCar("dir_listing/fixtures.car")
	file := fixture.MustGetNode("ą", "ę", "file-źł.txt")

	dnsLinks := dnslink.MustOpenDNSLink("dir_listing/dnslink.yml")
	dnsLink := dnsLinks.MustGet("dir-listing-website")

	tests := SugarTests{
		{
			Name: "Backlink on root CID should be hidden (TODO: cleanup Kubo-specifics)",
			Request: Request().
				Path("/").
				Header("Host", dnsLink),
			Response: Expect().
				Body(
					And(
						Contains("Index of"),
						Not(Contains(`<a href="/">..</a>`)),
					),
				),
		},
		{
			Name: "Redirect dir listing to URL with trailing slash",
			Request: Request().
				Path("/ą/ę").
				Header("Host", dnsLink),
			Response: Expect().
				Status(301).
				Headers(
					Header("Location").Equals(`/%C4%85/%C4%99/`),
				),
		},
		{
			Name: "Regular dir listing (TODO: cleanup Kubo-specifics)",
			Request: Request().
				Path("/ą/ę/").
				Header("Host", dnsLink),
			Response: Expect().
				Headers(
					Header("Etag").Contains(`"DirIndex-`),
					Header("Ipfs-Uri").
						Hint("DNSLink content paths produce an ipns:// URI with the dotted DNS name as authority (IPIP-548)").
						Spec("https://specs.ipfs.tech/http-gateways/path-gateway/#ipfs-uri-response-header").
						Equals("ipns://{{host}}{{path}}", dnsLink, "/%C4%85/%C4%99/"),
				).
				BodyWithHint(`
					- backlink on subdirectory should point at parent directory (TODO:  kubo-specific)
					- breadcrumbs should point at content root mounted at dnslink origin (TODO:  kubo-specific)
					- name column should be a link to content root mounted at dnslink origin
					- hash column should be a CID link to cid.ipfs.tech
					  DNSLink websites don't have public gateway mounted by default
					  See: https://github.com/ipfs/dir-index-html/issues/42 (TODO: class and other attrs are kubo-specific)
					`,
					And(
						Contains("Index of"),
						Contains(`<a href="/%C4%85/%C4%99/..">..</a>`),
						Contains(`/ipns/<a href="//{{hostname}}/">{{hostname}}</a>/<a href="//{{hostname}}/%C4%85">ą</a>/<a href="//{{hostname}}/%C4%85/%C4%99">ę</a>`, dnsLink),
						Contains(`<a href="/%C4%85/%C4%99/file-%C5%BA%C5%82.txt">file-źł.txt</a>`),
						Contains(`<a class="ipfs-hash" translate="no" href="https://cid.ipfs.tech/#{{cid}}" target="_blank" rel="noreferrer noopener">`, file.Cid()),
					),
				),
		},
	}

	RunWithSpecs(t, tests, specs.DNSLinkGateway)
}

func TestDNSLinkGatewayWithSubpath(t *testing.T) {
	tooling.LogTestGroup(t, GroupDNSLink)

	dnsLinks := dnslink.MustOpenDNSLink("gateway-cache/dnslink.yml")
	dnsLink := dnsLinks.MustGet("dnslink-with-subpath")

	tests := SugarTests{
		{
			Name: "GET for DNSLink with dnslink=/ipfs/<cid>/sub-path returns file at sub-path",
			Hint: `
			When a DNSLink TXT record points to a content path with a sub-path
			(e.g. /ipfs/<cid>/root2), the gateway must resolve the full path and
			concatenate any additional request path segments to serve the file.
			`,
			Request: Request().
				Header("Host", dnsLink).
				Path("/root3/root4/index.html"),
			Response: Expect().
				Status(200).
				Body("hello\n"),
		},
	}

	RunWithSpecs(t, tests, specs.DNSLinkGateway)
}
