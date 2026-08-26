package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/car"
	"github.com/ipfs/gateway-conformance/tooling/ipns"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/tmpl"
	"github.com/ipfs/go-cid"
	"github.com/mr-tron/base58/base58"
)

const (
	ipfsUriSpec   = "https://specs.ipfs.tech/http-gateways/path-gateway/#ipfs-uri-response-header"
	xIpfsPathSpec = "https://specs.ipfs.tech/http-gateways/path-gateway/#x-ipfs-path-response-header"
)

// TestGatewayIpfsUri asserts that responses come with an Ipfs-Uri header
// whose value is the one canonical serialization of the content path
// (IPIP-548): base32 CIDv1 authority, every path segment percent-encoded
// over the RFC 3986 unreserved set with uppercase hex, no query. Each
// request below pins a distinct failure class; the full byte-level test
// vectors live in the IPIP-0548 fixtures table.
func TestGatewayIpfsUri(t *testing.T) {
	tooling.LogTestGroup(t, GroupUnixFS)

	tricky := car.MustOpenUnixfsCar("path_gateway_unixfs/dir-with-tricky-filenames.car")
	trickyRoot := tricky.MustGetCid() // CIDv1 in canonical base32 form

	nested := car.MustOpenUnixfsCar("path_gateway_unixfs/dir-with-tricky-nested-filenames.car")
	nestedRoot := nested.MustGetCid()

	slashed := car.MustOpenUnixfsCar("path_gateway_unixfs/dir-with-slash-in-filename.car")
	slashedRoot := slashed.MustGetCid()

	fileCID := tricky.MustGetCid("plain.txt")

	// Legacy X-Ipfs-Path expectation for an ASCII-representable name: the
	// deprecated header is optional, but when present it carries the
	// decoded legacy content path unchanged.
	legacyPath := "/ipfs/" + trickyRoot + "/a#b?c.txt"

	tests := SugarTests{
		{
			Name: "GET for bare /ipfs/cid returns Ipfs-Uri without any path",
			Hint: "The URI path mirrors the content path remainder after /ipfs/{cid}: an empty remainder means no path in the URI",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipfs/{{cid}}", fileCID),
			Response: Expect().
				Status(200).
				Headers(
					Header("Ipfs-Uri").
						Equals("ipfs://{{cid}}", fileCID),
				),
		},
		{
			Name: "GET for nested /ipfs/ directory with trailing slash returns Ipfs-Uri with encoded segment and trailing slash",
			Hint: "A trailing slash on a nested directory is kept, and a directory segment is percent-encoded like any other; path cleanup helpers that strip non-root trailing slashes produce a different address",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipfs/{{cid}}/{{segment}}/", nestedRoot, "sub%20dir"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Ipfs-Uri").
						Equals("ipfs://{{cid}}/{{segment}}/", nestedRoot, "sub%20dir"),
				),
		},
		{
			Name: "GET for /ipfs/ file with sub-delims in the name returns Ipfs-Uri with every byte outside unreserved encoded",
			Hint: "encodeURIComponent leaves !'()* raw and URLEncoder over-encodes ~; the strict profile encodes every byte outside RFC 3986 unreserved (uppercase hex) and nothing else",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipfs/{{cid}}/{{dir}}/{{file}}", nestedRoot, "sub%20dir", "file%21%27%28%29%2A~.txt"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Ipfs-Uri").
						Equals("ipfs://{{cid}}/{{dir}}/{{file}}", nestedRoot, "sub%20dir", "file%21%27%28%29%2A~.txt"),
				).
				Body(nested.MustGetRawData("sub dir", "file!'()*~.txt")),
		},
		{
			Name: "GET for /ipfs/ file with encoded slash and dot segments in request path returns Ipfs-Uri with the resolved path",
			Hint: "The content path is derived by percent-decoding each request path component once (%2F becomes a separator, %2E a dot) and then resolving dot segments; Ipfs-Uri describes the path the gateway actually resolved and served",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipfs/{{cid}}/{{segment}}", trickyRoot, "a%2F..%2Fplain%2Etxt"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Ipfs-Uri").
						Equals("ipfs://{{cid}}/plain.txt", trickyRoot),
				).
				Body(tricky.MustGetRawData("plain.txt")),
		},
		{
			Name: "GET for /ipfs/ nested file whose parent also has a link with / in its name returns the nested file",
			Hint: "A dag-pb link name may contain a slash at the byte level, but path components never contain /: the path a/b.txt always means b.txt inside the a subdirectory, never the sibling link literally named 'a/b.txt'",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipfs/{{cid}}/a/b.txt", slashedRoot),
			Response: Expect().
				Status(200).
				Headers(
					Header("Ipfs-Uri").
						Equals("ipfs://{{cid}}/a/b.txt", slashedRoot),
				).
				Body(slashed.MustGetRawData("a", "b.txt")),
		},
		{
			Name: "GET for /ipfs/ nested file via %2F returns the nested file, never the link with / in its name",
			Hint: "%2F decodes to a path separator, so a%2Fb.txt is the same content path as a/b.txt; an implementation that instead matches the raw segment against link names would serve the sibling link literally named 'a/b.txt', which is not addressable by any content path",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipfs/{{cid}}/{{segment}}", slashedRoot, "a%2Fb.txt"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Ipfs-Uri").
						Equals("ipfs://{{cid}}/a/b.txt", slashedRoot),
				).
				Body(slashed.MustGetRawData("a", "b.txt")),
		},
		{
			Name: "GET for /ipfs/ file with filename and format query parameters returns Ipfs-Uri without the query",
			Hint: "Request query parameters are never included in Ipfs-Uri, and alternate response formats set the header too",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipfs/{{cid}}/plain.txt", trickyRoot).
				Query("filename", "download.txt").
				Query("format", "json"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Ipfs-Uri").
						Equals("ipfs://{{cid}}/plain.txt", trickyRoot),
				),
		},
		{
			Name: "GET for /ipfs/ file with URI delimiters in the name returns encoded Ipfs-Uri and sane X-Ipfs-Path",
			Hint: "# and ? in a file name are percent-encoded so the value stays one URI; the deprecated X-Ipfs-Path header is optional, but when present its value is the decoded legacy content path, unchanged",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipfs/{{cid}}/{{segment}}", trickyRoot, "a%23b%3Fc.txt"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Ipfs-Uri").
						Equals("ipfs://{{cid}}/{{segment}}", trickyRoot, "a%23b%3Fc.txt"),
					Header("X-Ipfs-Path").
						Hint(Fmt("when present, the value must be exactly '{{path}}'", legacyPath)).
						Spec(xIpfsPathSpec).
						ChecksAll(func(values []string) bool {
							return len(values) == 0 || (len(values) == 1 && values[0] == legacyPath)
						}),
				).
				Body(tricky.MustGetRawData("a#b?c.txt")),
		},
		{
			Name: "GET for /ipfs/ file with non-ASCII name returns percent-encoded Ipfs-Uri and no X-Ipfs-Path",
			Hint: "Non-ASCII names are percent-encoded over their UTF-8 bytes (a 4-byte emoji covers UTF-16 surrogate pairs); the deprecated X-Ipfs-Path header MUST be omitted, since bytes outside HTAB, SP, and visible ASCII cannot survive an HTTP field value (RFC 9110, Section 5.5)",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipfs/{{cid}}/{{segment}}", trickyRoot, "emoji%F0%9F%9A%80.txt"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Ipfs-Uri").
						Equals("ipfs://{{cid}}/{{segment}}", trickyRoot, "emoji%F0%9F%9A%80.txt"),
					Header("X-Ipfs-Path").
						Hint("the deprecated header must be omitted when the content path cannot appear in an HTTP field value").
						Spec(xIpfsPathSpec).
						IsEmpty(),
				).
				Body(tricky.MustGetRawData("emoji🚀.txt")),
		},
		{
			Name: "GET for /ipfs/ directory without trailing slash returns redirect with Ipfs-Uri",
			Hint: "The header SHOULD also be returned on redirect responses once the content root has been parsed; it indicates the requested content path, before the trailing-slash normalization the redirect performs",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipfs/{{cid}}", trickyRoot),
			Response: Expect().
				Status(301).
				Headers(
					Header("Location").
						Equals("/ipfs/{{cid}}/", trickyRoot),
					Header("Ipfs-Uri").
						Equals("ipfs://{{cid}}", trickyRoot),
				),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)
}

// TestGatewayIpfsUriPercentEncodedFilename asserts that the Ipfs-Uri value
// is the canonical serialization of the content path no matter how the
// request spelled it. The fixture name contains a literal percent-triple,
// sub-delims (+, =), a space, and non-ASCII characters: implementations
// that echo the raw request path, decode it twice, or keep lowercase hex
// all produce a different value. Same fixture and expected value as the
// IPIP-0548 test fixtures table.
func TestGatewayIpfsUriPercentEncodedFilename(t *testing.T) {
	tooling.LogTestGroup(t, GroupUnixFS)

	fixture := car.MustOpenUnixfsCar("path_gateway_unixfs/dir-with-percent-encoded-filename.car")
	rootCID := fixture.MustGetCid()

	const (
		fileName  = "Portugal%2C+España=Peninsula Ibérica.txt"
		canonical = "Portugal%252C%2BEspa%C3%B1a%3DPeninsula%20Ib%C3%A9rica.txt"
	)

	// Different request spellings of the same content path: each decodes
	// once to the file name above, so each must produce the same
	// canonical Ipfs-Uri value.
	spellings := []struct {
		kind    string
		segment string
	}{
		{"canonical spelling", canonical},
		{"raw sub-delims spelling", "Portugal%252C+Espa%C3%B1a=Peninsula%20Ib%C3%A9rica.txt"},
		// Lowercase hex only in the real escapes: the 2C after %25 is
		// literal name content and keeps its case. The trailing %2etxt
		// also proves an unreserved escape is decoded, not passed through.
		{"lowercase hex spelling", "Portugal%252C%2bEspa%c3%b1a%3dPeninsula%20Ib%c3%a9rica%2etxt"},
	}

	var tests SugarTests
	for i, s := range spellings {
		headers := []HeaderBuilder{
			Header("Ipfs-Uri").
				Equals("ipfs://{{cid}}/{{segment}}", rootCID, canonical),
		}
		if i == 0 {
			headers = append(headers,
				Header("X-Ipfs-Path").
					Hint("the deprecated header must be omitted when the content path cannot appear in an HTTP field value").
					Spec(xIpfsPathSpec).
					IsEmpty(),
			)
		}
		tests = append(tests, SugarTest{
			Name: Fmt("GET for /ipfs/ percent-encoded-looking file via {{kind}} returns canonical Ipfs-Uri", s.kind),
			Hint: "The Ipfs-Uri value depends only on the content path, never on how the request spelled it: each request segment is decoded once, then re-encoded with the strict profile (uppercase hex, sub-delims encoded, the literal %2C in the name encoded as %252C)",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipfs/{{cid}}/{{segment}}", rootCID, s.segment),
			Response: Expect().
				Status(200).
				Headers(headers...).
				Body(fixture.MustGetRawData(fileName)),
		})
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)
}

// TestGatewayIpfsUriCidNormalization asserts that the authority of the
// Ipfs-Uri value is the canonical form of the content root: a CIDv0
// request is normalized to the same CID as a CIDv1 in lowercase base32.
func TestGatewayIpfsUriCidNormalization(t *testing.T) {
	tooling.LogTestGroup(t, GroupUnixFS)

	fixture := car.MustOpenUnixfsCar("path_gateway_unixfs/dir-with-tricky-filenames.car")
	rootCIDv1 := fixture.MustGetCid() // CIDv1 in canonical base32 form

	// The root is dag-pb with a sha2-256 hash, so the same CID also has a
	// legacy CIDv0 (Qm...) spelling.
	parsed := cid.MustParse(rootCIDv1)
	rootCIDv0 := cid.NewCidV0(parsed.Hash()).String()

	tests := SugarTests{
		{
			Name: "GET for /ipfs/ with CIDv0 returns Ipfs-Uri with CID normalized to base32 CIDv1",
			Hint: "The URI authority is always the canonical base32 CIDv1 form, no matter how the request encoded the CID",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipfs/{{cid}}/plain.txt", rootCIDv0),
			Response: Expect().
				Status(200).
				Headers(
					Header("Ipfs-Uri").
						Equals("ipfs://{{cid}}/plain.txt", rootCIDv1),
				),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)
}

// TestGatewayIpfsUriWithIPNS asserts that content paths in the /ipns/
// namespace produce an ipns:// URI with a canonical base36 libp2p-key
// authority, no matter whether the request spelled the name in base36 or
// as a legacy base58 peer ID.
func TestGatewayIpfsUriWithIPNS(t *testing.T) {
	tooling.LogTestGroup(t, GroupIPNS)

	record := ipns.MustOpenIPNSRecordWithKey("gateway-cache/k51qzi5uqu5dlxdsdu5fpuu7h69wu4ohp32iwm9pdt9nq3y5rpn3ln9j12zfhe.ipns-record")
	ipnsKey := record.Key() // base36 (k...) libp2p-key CIDv1

	// The same key as a legacy base58btc peer ID string (12D3Koo...):
	// the raw multihash of the libp2p-key CID.
	peerID := base58.Encode(cid.MustParse(ipnsKey).Hash())

	canonicalIpnsURI := "ipns://" + ipnsKey + "/root2/root3/root4/index.html"
	canonicalIpnsPath := "/ipns/" + ipnsKey + "/root2/root3/root4/index.html"

	tests := SugarTests{
		{
			Name: "GET for nested /ipns/ directory returns Ipfs-Uri with ipns:// URI, base36 authority, and trailing slash",
			Hint: "Cryptographic IPNS names use the canonical base36 libp2p-key CIDv1 form as the URI authority, and a trailing slash on a nested directory is kept",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipns/{{key}}/root2/root3/", ipnsKey),
			Response: Expect().
				Status(200).
				Headers(
					Header("Ipfs-Uri").
						Equals("ipns://{{key}}/root2/root3/", ipnsKey),
				),
		},
		{
			Name: "GET for /ipns/ via legacy base58 peer ID normalizes the name to base36",
			Hint: "A legacy base58 peer ID under /ipns/ becomes a libp2p-key CIDv1 in base36; conversion helpers that fall back to the CID default of base32 produce a different name. A gateway MAY answer with a canonicalization redirect instead of serving the path directly; either way the base36 form is the only acceptable authority",
			Spec: ipfsUriSpec,
			Request: Request().
				Path("/ipns/{{peerid}}/root2/root3/root4/index.html", peerID),
			Response: Expect().
				StatusBetween(200, 399).
				Headers(
					Header("Ipfs-Uri").
						Hint(Fmt("when present, the value must carry the base36 authority: '{{uri}}'", canonicalIpnsURI)).
						ChecksAll(func(values []string) bool {
							return len(values) == 0 || (len(values) == 1 && values[0] == canonicalIpnsURI)
						}),
					Header("Location").
						Hint("a canonicalization redirect, when used, must point at the base36 form").
						ChecksAll(func(values []string) bool {
							return len(values) == 0 || (len(values) == 1 && values[0] == canonicalIpnsPath)
						}),
				),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS, specs.PathGatewayIPNS)
}
