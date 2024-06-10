package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/tmpl"
)

var (
	fixture = car.MustOpenUnixfsCar("subdomain_gateway/fixtures.car")

	CIDVal   = string(fixture.MustGetRawData("hello-CIDv1")) // hello
	DirCID   = fixture.MustGetCid("testdirlisting")
	CIDv1    = fixture.MustGetCid("hello-CIDv1")
	CIDv0    = fixture.MustGetCid("hello-CIDv0")
	CIDv0to1 = fixture.MustGetCid("hello-CIDv0to1")
	//CIDv1_TOO_LONG = fixture.MustGetCid("hello-CIDv1_TOO_LONG")

	// the gateway endpoint is used as HTTP proxy
	gatewayAsProxyURL = GatewayURL().String()

	// run against origins explicitly passed via --subdomain-url
	s = SubdomainGatewayURL()
)

func TestProxyGatewaySubdomains(t *testing.T) {
	tests := SugarTests{
		{
			Name: "request for {CID}.ipfs.example.com should return expected payload",
			Hint: "HTTP proxy gateway accepts requests for GETs of full URLs as Paths",
			Request: Request().
				Proxy(gatewayAsProxyURL).
				Path("{{scheme}}://{{cid}}.ipfs.{{host}}", s.Scheme, CIDv1, s.Host),
			Response: Expect().
				Status(200).
				Body(Contains(CIDVal)),
		},
		{
			Name: "request for example.com/ipfs/{CIDv0} redirects to {CIDv1}.ipfs.example.com",
			Hint: "HTTP proxy gateway accepts requests for GETs of full URLs as Paths",
			Request: Request().
				Proxy(gatewayAsProxyURL).
				Path("{{scheme}}://{{host}}/ipfs/{{cid}}/", s.Scheme, s.Host, CIDv0),
			Response: Expect().
				Status(301).
				Headers(
					Header("Location").
						Hint("request for example.com/ipfs/{CIDv0to1} returns Location HTTP header for subdomain redirect in browsers").
						Contains("{{scheme}}://{{cid}}.ipfs.{{host}}/", s.Scheme, CIDv0to1, s.Host),
				),
		},
		{
			Name: "request for {CID}.ipfs.example.com/ipfs/file.txt should return data from a file in CID content root",
			Hint: "ensure subdomain gateway takes priority over processing /ipfs/* paths",
			Request: Request().
				Proxy(gatewayAsProxyURL).
				Path("{{scheme}}://{{cid}}.ipfs.{{host}}/ipfs/file.txt", s.Scheme, DirCID, s.Host),
			Response: Expect().
				Status(200).
				Body(Contains("I am a txt file")),
		},
		/* TODO: value added
		{
			Name: "request for a too long CID at {CIDv1}.ipfs.example.com returns expected payload",
			Hint: "HTTP proxy mode allows responding to requests with 'DNS labels' longer than 63 characters",
			Request: Request().
				Proxy(gatewayAsProxyURL).
				TODO turn to Path: Header("Host", Fmt("{{cid}}.ipfs.{{host}}", CIDv1_TOO_LONG, s.Host)).
				Path("/"),
			Response: Expect().
				Status(400).
				Body(Contains("TODO")),
		},
		*/
	}
	RunWithSpecs(t, tests, specs.ProxyGateway, specs.SubdomainGatewayIPFS)
}

func TestProxyTunnelGatewaySubdomains(t *testing.T) {
	tests := SugarTests{
		{
			Name: "request for {CID}.ipfs.example.com should return expected payload",
			Hint: "HTTP CONNECT is how some proxy setups convert an HTTP connection into a tunnel to a remote host https://tools.ietf.org/html/rfc7231#section-4.3.6",
			Request: Request().
				WithProxyTunnel().
				Proxy(gatewayAsProxyURL).
				Header("Host", Fmt("{{cid}}.ipfs.{{host}}", CIDv1, s.Host)).
				Path("/"),
			Response: Expect().
				Status(200).
				Body(Contains(CIDVal)),
		},
		{
			Name: "request for example.com/ipfs/{CIDv0} redirects to {CIDv1}.ipfs.example.com",
			Hint: "proxy tunnel follows ",
			Request: Request().
				WithProxyTunnel().
				Proxy(gatewayAsProxyURL).
				Header("Host", s.Host).
				Path("/ipfs/{{cid}}/", CIDv0),
			Response: Expect().
				Status(301).
				Headers(
					Header("Location").
						Hint("request for example.com/ipfs/{CIDv0to1} returns Location HTTP header for subdomain redirect in browsers").
						Contains("{{scheme}}://{{cid}}.ipfs.{{host}}/", s.Scheme, CIDv0to1, s.Host),
				),
		},
		{
			Name: "request for {CID}.ipfs.example.com/ipfs/file.txt should return data from a file in CID content root",
			Hint: "ensure subdomain gateway takes priority over processing /ipfs/* paths",
			Request: Request().
				WithProxyTunnel().
				Proxy(gatewayAsProxyURL).
				Header("Host", Fmt("{{cid}}.ipfs.{{host}}", DirCID, s.Host)).
				Path("/ipfs/file.txt"),
			Response: Expect().
				Status(200).
				Body(Contains("I am a txt file")),
		},
		/* TODO: value added
		{
			Name: "request for a too long CID at {CIDv1}.ipfs.example.com returns expected payload",
			Hint: "HTTP proxy mode allows responding to requests with 'DNS labels' longer than 63 characters",
			Request: Request().
				WithProxyTunnel().
				Proxy(gatewayAsProxyURL).
				TODO turn to Path: Header("Host", Fmt("{{cid}}.ipfs.{{host}}", CIDv1_TOO_LONG, s.Host)).
				Path("/"),
			Response: Expect().
				Status(400).
				Body(Contains("TODO")),
		},
		*/
	}
	RunWithSpecs(t, tests, specs.ProxyGateway, specs.SubdomainGatewayIPFS)
}
