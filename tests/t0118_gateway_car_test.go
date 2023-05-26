package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewayCar(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0118-test-dag.car")
	hamtFixture := car.MustOpenUnixfsCar("t0118-ipip-402-scope.car")

	dagScopeAllBuilder := Expect().
		Status(200).
		Body(
			IsCar().
			HasRoot(hamtFixture.MustGetCid()).
			HasBlocks(hamtFixture.MustGetCid()).
			HasBlocks(
				hamtFixture.MustGetCid("sub2"),
				hamtFixture.MustGetCid("sub2", "hello.txt"),
			).
			HasBlocks(hamtFixture.MustGetChildrenCids("sub2", "hello.txt")...).
			HasBlocks(
				hamtFixture.MustGetCid("sub1"),
				hamtFixture.MustGetCid("sub1", "hello.txt"),
			).
			HasBlocks(hamtFixture.MustGetChildrenCids("sub1", "hello.txt")...).
			Exactly().
			InThatOrder(),
		)

	tests := SugarTests{
		{
			Name: "GET response for application/vnd.ipld.car",
			Hint: `
				CAR stream is not deterministic, as blocks can arrive in random order,
				but if we have a small file that fits into a single block, and export its CID
				we will get a CAR that is a deterministic array of bytes.
			`,
			Request: Request().
				Path("ipfs/{{cid}}/subdir/ascii.txt", fixture.MustGetCid()).
				Headers(
					Header("Accept", "application/vnd.ipld.car"),
				),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").
						Hint("Expected content type to be application/vnd.ipld.car").
						Contains("application/vnd.ipld.car"),
					Header("Content-Length").
						Hint("CAR is streamed, gateway may not have the entire thing, unable to calculate total size").
						IsEmpty(),
					Header("Content-Disposition").
						Hint(`Expected content disposition to be attachment; filename="<cid>.car"`).
						Contains(`attachment; filename="{{cid}}.car"`, fixture.MustGetCid("subdir", "ascii.txt")),
					Header("X-Content-Type-Options").
						Hint("CAR is streamed, gateway may not have the entire thing, unable to calculate total size").
						Equals("nosniff"),
					Header("Accept-Ranges").
						Hint("CAR is streamed, gateway may not have the entire thing, unable to support range-requests. Partial downloads and resumes should be handled using IPLD selectors: https://github.com/ipfs/go-ipfs/issues/8769").
						Equals("none"),
				).
				Body(
					IsCar().
					HasRoot(fixture.MustGetCid()).
					HasBlocks(
						fixture.MustGetCid(),
						fixture.MustGetCid("subdir"),
						fixture.MustGetCid("subdir", "ascii.txt"),
					).
					Exactly().
					InThatOrder(),
				),
		},
		{
			Name: "GET with ?format=car&dag-scope=block params returns expected blocks",
			Hint: `
				dag-scope=block should return a CAR file with only the root block and a
				block for each optional path component.
			`,
			Request: Request().
				Path("ipfs/{{cid}}/sub1/hello.txt", hamtFixture.MustGetCid()).
				Query("format", "car").
				Query("dag-scope", "block"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
					HasRoot(hamtFixture.MustGetCid()).
					HasBlocks(
						hamtFixture.MustGetCid(),
						hamtFixture.MustGetCid("sub1"),
						hamtFixture.MustGetCid("sub1", "hello.txt"),
					).
					Exactly().
					InThatOrder(),
				),
		},
		{
			Name: "GET with ?format=car&dag-scope=entity params returns expected blocks",
			Hint: `
				dag-scope=entity should return a CAR file with all the blocks needed to 'cat'
				a UnixFS file at the end of the specified path, or to 'ls' a UnixFS directory
				at the end of the specified path.
			`,
			Request: Request().
				Path("ipfs/{{cid}}/sub1/hello.txt", hamtFixture.MustGetCid()).
				Query("format", "car").
				Query("dag-scope", "entity"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
					HasRoot(hamtFixture.MustGetCid()).
					HasBlocks(
						hamtFixture.MustGetCid(),
						hamtFixture.MustGetCid("sub1"),
						hamtFixture.MustGetCid("sub1", "hello.txt"),
					).
					HasBlocks(hamtFixture.MustGetChildrenCids("sub1", "hello.txt")...).
					Exactly().
					InThatOrder(),
				),
		},
		{
			Name: "GET with ?format=car&dag-scope=all params returns expected blocks",
			Hint: `
				dag-scope=all should return a CAR file with the entire contiguous DAG
				that begins at the end of the path query, after blocks required to verify path segments.
			`,
			Request: Request().
				Path("ipfs/{{cid}}", hamtFixture.MustGetCid()).
				Query("format", "car").
				Query("dag-scope", "all"),
			Response: dagScopeAllBuilder,
		},
		{
			Name: "GET with ?format=car returns same response as dag-scope=all",
			Hint: `If the CAR request does not have a dag-scope parameter, it should be treated as dag-scope=all`,
			Request: Request().
				Path("ipfs/{{cid}}", hamtFixture.MustGetCid()).
				Query("format", "car"),
			Response: dagScopeAllBuilder,
		},
	}

	Run(t, tests)
}
