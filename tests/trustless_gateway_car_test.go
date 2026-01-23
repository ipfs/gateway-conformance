package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/helpers"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestTrustlessCarPathing(t *testing.T) {
	tooling.LogTestGroup(t, GroupBlockCar)

	subdirTwoSingleBlockFilesFixture := car.MustOpenUnixfsCar("trustless_gateway_car/subdir-with-two-single-block-files.car")
	singleLayerHamtMultiBlockFilesFixture := car.MustOpenUnixfsCar("trustless_gateway_car/single-layer-hamt-with-multi-block-files.car")
	dirWithDagCborWithLinksFixture := car.MustOpenUnixfsCar("trustless_gateway_car/dir-with-dag-cbor-with-links.car")

	tests := SugarTests{
		{
			Name: "GET default CAR response with pathing through UnixFS Directory",
			Hint: `
				CAR stream of a UnixFS file within a path under UnixFS subdirectory should contain
				all the blocks to traverse the path, as well as all the blocks behind
				the last path segment (terminating entity returned recursively).
			`,
			Request: Request().
				Path("/ipfs/{{cid}}/subdir/ascii.txt", subdirTwoSingleBlockFilesFixture.MustGetCidWithCodec(0x70)).
				Query("format", "car"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							subdirTwoSingleBlockFilesFixture.MustGetCid(),
							subdirTwoSingleBlockFilesFixture.MustGetCid("subdir"),
							subdirTwoSingleBlockFilesFixture.MustGetCid("subdir", "ascii.txt"),
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET default CAR response of UnixFS file on a path with HAMT-sharded directory",
			Hint: `
				CAR stream of a UnixFS file within a path with sharded directory should contain
				all the blocks to traverse the path, as well as all the blocks contained inside
				the last path segment, recursively.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}/685.txt", singleLayerHamtMultiBlockFilesFixture.MustGetCidWithCodec(0x70)).
				Query("format", "car"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(flattenStrings(t,
							singleLayerHamtMultiBlockFilesFixture.MustGetCid(),
							singleLayerHamtMultiBlockFilesFixture.MustGetCIDsInHAMTTraversal(nil, "685.txt"),
							singleLayerHamtMultiBlockFilesFixture.MustGetCid("685.txt"),
							singleLayerHamtMultiBlockFilesFixture.MustGetDescendantsCids("685.txt"))...,
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET default CAR response of UnixFS file on a path with DAG-CBOR as root CID",
			Hint: `
				CAR stream of a UnixFS file on a path with DAG-CBOR as root CID resolves IPLD Link
				and returns all the blocks to traverse the path, as well as all the blocks contained
				inside the last path segment, recursively.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}/files/single", dirWithDagCborWithLinksFixture.MustGetCidWithCodec(0x71, "document")).
				Query("format", "car"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(flattenStrings(t,
							dirWithDagCborWithLinksFixture.MustGetCid("document"),
							dirWithDagCborWithLinksFixture.MustGetCid("document", "files", "single"),
						)...).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET default CAR response for non-existing file",
			Hint: `
				The response code depends on implementation details such as the locality and the cost of path traversal checks,
				and trade-off between latency and correctness.
				Implementations that are able to efficiently detect requested content path does not exist,
				should not return CAR response, but a simple 404.
				Implementations that are focusing on stateless streaming and low latency are free to return
				partial CAR up to the missing link (blocks necessary to traverse the path up to and including
				the parent of the first non-existing segment).
			`,
			Request: Request().
				Path("/ipfs/{{cid}}/subdir/i-do-not-exist", subdirTwoSingleBlockFilesFixture.MustGetCidWithCodec(0x70)).
				Query("format", "car"),
			Response: AnyOf(
				// Stateless streaming implementations: 200 with partial CAR
				Expect().
					Status(200).
					Body(
						IsCar().
							IgnoreRoots().
							HasBlocks(
								subdirTwoSingleBlockFilesFixture.MustGetCid(),
								subdirTwoSingleBlockFilesFixture.MustGetCid("subdir"),
							).
							Exactly().
							InThatOrder(),
					),
				// Implementations with efficient path checks: 404 Not Found
				Expect().
					Status(404),
				// TODO: remove once Kubo ships with https://github.com/ipfs/boxo/pull/1019
				// Legacy behavior: 200 with X-Stream-Error header when missing blocks detected during streaming
				// is returned by boxo/gateway with remote car backend that implements the above PR
				Expect().
					Status(200).
					Headers(
						Header("X-Stream-Error").Not().IsEmpty(),
					),
			),
		},
	}

	RunWithSpecs(t, helpers.StandardCARTestTransforms(t, tests), specs.TrustlessGatewayCAR)
}

func TestTrustlessCarDagScopeBlock(t *testing.T) {
	tooling.LogTestGroup(t, GroupBlockCar)

	subdirTwoSingleBlockFilesFixture := car.MustOpenUnixfsCar("trustless_gateway_car/subdir-with-two-single-block-files.car")
	singleLayerHamtMultiBlockFilesFixture := car.MustOpenUnixfsCar("trustless_gateway_car/single-layer-hamt-with-multi-block-files.car")

	tests := SugarTests{
		{
			Name: "GET CAR with dag-scope=block of UnixFS directory on a path",
			Hint: `
				dag-scope=block should return a CAR file with only the root block at the
				end of the path and blocks required to verify the specified path segments.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}/subdir", subdirTwoSingleBlockFilesFixture.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "block"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							subdirTwoSingleBlockFilesFixture.MustGetCid(),
							subdirTwoSingleBlockFilesFixture.MustGetCid("subdir"),
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with dag-scope=block of UnixFS file on a path",
			Hint: `
				dag-scope=block should return a CAR file with only the root block at the
				end of the path and blocks required to verify the specified path segments.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}/subdir/ascii.txt", subdirTwoSingleBlockFilesFixture.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "block"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							subdirTwoSingleBlockFilesFixture.MustGetCid(),
							subdirTwoSingleBlockFilesFixture.MustGetCid("subdir"),
							subdirTwoSingleBlockFilesFixture.MustGetCid("subdir", "ascii.txt"),
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with dag-scope=block of UnixFS file on a path with sharded directory",
			Hint: `
				dag-scope=block should return a CAR file with only the root block at the
				end of the path and blocks required to verify the specified path segments.
				Pathing through a sharded directory should return the blocks needed for the
				traversal, not the entire HAMT and not skipping all intermediate nodes.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}/1.txt", singleLayerHamtMultiBlockFilesFixture.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "block"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(flattenStrings(t,
							singleLayerHamtMultiBlockFilesFixture.MustGetCid(),
							singleLayerHamtMultiBlockFilesFixture.MustGetCIDsInHAMTTraversal(nil, "1.txt"),
							singleLayerHamtMultiBlockFilesFixture.MustGetCid("1.txt"))...,
						).
						Exactly().
						InThatOrder(),
				),
		},
	}

	RunWithSpecs(t, helpers.StandardCARTestTransforms(t, tests), specs.TrustlessGatewayCAR)
}

func TestTrustlessCarDagScopeEntity(t *testing.T) {
	tooling.LogTestGroup(t, GroupBlockCar)

	subdirTwoSingleBlockFilesFixture := car.MustOpenUnixfsCar("trustless_gateway_car/subdir-with-two-single-block-files.car")
	singleLayerHamtMultiBlockFilesFixture := car.MustOpenUnixfsCar("trustless_gateway_car/single-layer-hamt-with-multi-block-files.car")
	subdirWithMixedBlockFiles := car.MustOpenUnixfsCar("trustless_gateway_car/subdir-with-mixed-block-files.car")
	dirWithDagCborWithLinksFixture := car.MustOpenUnixfsCar("trustless_gateway_car/dir-with-dag-cbor-with-links.car")

	tests := SugarTests{
		{
			Name: "GET CAR with dag-scope=entity of a UnixFS directory",
			Hint: `
				dag-scope=entity for a directory should return a CAR file with all of the path blocks, as well
				as all of the blocks for directory enumeration, but not any of blocks for items below the directory.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", subdirTwoSingleBlockFilesFixture.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "entity"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							subdirTwoSingleBlockFilesFixture.MustGetCid(),
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with dag-scope=entity of a UnixFS sharded directory",
			Hint: `
				dag-scope=entity for a sharded directory should return a CAR file with all of the path blocks as well
				as all of the blocks in the HAMT, but not any of blocks below the HAMT.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", singleLayerHamtMultiBlockFilesFixture.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "entity"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							flattenStrings(t,
								singleLayerHamtMultiBlockFilesFixture.MustGetCid(),
								singleLayerHamtMultiBlockFilesFixture.MustGetCidsInHAMT())...,
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with dag-scope=entity of a UnixFS file",
			Hint: `
				dag-scope=entity for a UnixFS file within a directory must return all necessary
				blocks to verify the path, as well as to decode the full file.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}/subdir/ascii.txt", subdirWithMixedBlockFiles.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "entity"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid(),
								subdirWithMixedBlockFiles.MustGetCid("subdir"),
								subdirWithMixedBlockFiles.MustGetCid("subdir", "ascii.txt"),
							)...,
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with dag-scope=entity of a chunked UnixFS file",
			Hint: `
				dag-scope=entity for a chunked UnixFS file within a directory must return
				all necessary blocks to verify the path, as well as to decode the full file.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}/subdir/multiblock.txt", subdirWithMixedBlockFiles.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "entity"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid(),
								subdirWithMixedBlockFiles.MustGetCid("subdir"),
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetDescendantsCids("subdir", "multiblock.txt"),
							)...,
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with dag-scope=entity of DAG-CBOR with Links",
			Hint: `
				dag-scope=entity of a DAG-CBOR (or DAG-JSON) document with IPLD Links must return
				all necessary blocks to verify the path, the document itself, but not the content of the IPLD Links.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}/document", dirWithDagCborWithLinksFixture.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "entity"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(flattenStrings(t,
							dirWithDagCborWithLinksFixture.MustGetCid(),
							dirWithDagCborWithLinksFixture.MustGetCid("document"),
						)...).
						Exactly().
						InThatOrder(),
				),
		},
	}

	RunWithSpecs(t, helpers.StandardCARTestTransforms(t, tests), specs.TrustlessGatewayCAR)
}

func TestTrustlessCarDagScopeAll(t *testing.T) {
	tooling.LogTestGroup(t, GroupBlockCar)

	subdirWithMixedBlockFiles := car.MustOpenUnixfsCar("trustless_gateway_car/subdir-with-mixed-block-files.car")

	tests := SugarTests{
		{
			Name: "GET CAR with dag-scope=all of UnixFS directory with multiple files",
			Hint: `
				dag-scope=all should return a blocks required to verify path, and then 
				all blocks for the entire UnixFS directory DAG (all children, recursively).
			`,
			Request: Request().
				Path("/ipfs/{{cid}}/subdir", subdirWithMixedBlockFiles.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "all"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							flattenStrings(
								t,
								subdirWithMixedBlockFiles.MustGetCid(),
								subdirWithMixedBlockFiles.MustGetCid("subdir"),
								subdirWithMixedBlockFiles.MustGetDescendantsCids("subdir"),
							)...,
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with dag-scope=all of a chunked UnixFS file",
			Hint: `
				dag-scope=all for a chunked UnixFS file within a directory must return
				all necessary blocks to verify the path, as well as to decode the full file.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}/subdir/multiblock.txt", subdirWithMixedBlockFiles.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "all"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid(),
								subdirWithMixedBlockFiles.MustGetCid("subdir"),
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetDescendantsCids("subdir", "multiblock.txt"),
							)...,
						).
						Exactly().
						InThatOrder(),
				),
		},
	}

	RunWithSpecs(t, helpers.StandardCARTestTransforms(t, tests), specs.TrustlessGatewayCAR)
}

func TestTrustlessCarEntityBytes(t *testing.T) {
	tooling.LogTestGroup(t, GroupBlockCar)
	tooling.LogSpecs(t, "https://specs.ipfs.tech/http-gateways/trustless-gateway/#entity-bytes-request-query-parameter")

	singleLayerHamtMultiBlockFilesFixture := car.MustOpenUnixfsCar("trustless_gateway_car/single-layer-hamt-with-multi-block-files.car")
	subdirWithMixedBlockFiles := car.MustOpenUnixfsCar("trustless_gateway_car/subdir-with-mixed-block-files.car")
	missingBlockFixture := car.MustOpenUnixfsCar("trustless_gateway_car/file-3k-and-3-blocks-missing-block.car")

	tests := SugarTests{
		{
			Name: "GET CAR with entity-bytes succeeds even if the gateway is missing a block after the requested range",
			Hint: `
				dag-scope=entity&entity-bytes=0:x should return a CAR file with
				only the blocks needed to fullfill the request. This MUST
				succeed despite the fact that bytes beyond 'x' are not retrievable.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", missingBlockFixture.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "entity").
				Query("entity-bytes", "0:1000"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							missingBlockFixture.MustGetCid(),
							// This CID is defined at the SPEC level
							// See the recipe for `file-3k-and-3-blocks-missing-block.car`
							"QmPKt7ptM2ZYSGPUc8PmPT2VBkLDK3iqpG9TBJY7PCE9rF",
						).
						Exactly(),
				),
		},
		{
			Name: "GET CAR with entity-bytes succeeds even if the gateway is missing a block before the requested range",
			Hint: `
				dag-scope=entity&entity-bytes=y:* should return a CAR file with
				only the blocks needed to fullfill the request. This MUST
				succeed despite the fact that a block with bytes before 'y' is
				not retrievable.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", missingBlockFixture.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "entity").
				Query("entity-bytes", "2200:*"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							missingBlockFixture.MustGetCid(),
							// This CID is defined at the SPEC level
							// See the recipe for `file-3k-and-3-blocks-missing-block.car`
							"QmWXY482zQdwecnfBsj78poUUuPXvyw2JAFAEMw4tzTavV",
						).
						Exactly(),
				),
		},
		{
			Name: "GET CAR with entity-bytes of a full UnixFS file",
			Hint: `
				dag-scope=entity&entity-bytes=0:* should return a CAR file with all the blocks needed to 'cat'
				the full UnixFS file at the end of the specified path
			`,
			Request: Request().
				Path("/ipfs/{{cid}}/subdir/multiblock.txt", subdirWithMixedBlockFiles.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "entity").
				Query("entity-bytes", "0:*"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(flattenStrings(t,
							subdirWithMixedBlockFiles.MustGetCid(),
							subdirWithMixedBlockFiles.MustGetCid("subdir"),
							subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
							subdirWithMixedBlockFiles.MustGetDescendantsCids("subdir", "multiblock.txt"),
						)...).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with entity-bytes of a UnixFS directory",
			Hint: `
				dag-scope=entity&entity-bytes=from:to should return a CAR file with all the blocks needed to enumerate contents of
				a UnixFS directory at the end of the specified path if the terminal element is a directory
				(i.e. entity-bytes is effectively optional if the entity is not a file)
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", singleLayerHamtMultiBlockFilesFixture.MustGetCidWithCodec(0x70)).
				Query("format", "car").
				Query("dag-scope", "entity").
				Query("entity-bytes", "0:*"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							flattenStrings(t,
								singleLayerHamtMultiBlockFilesFixture.MustGetCid(),
								singleLayerHamtMultiBlockFilesFixture.MustGetCidsInHAMT())...,
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with entity-bytes equivalent to a HTTP Range Request from the middle of a file to the end",
			Hint: `
				The response MUST contain only the minimal set of blocks necessary for fulfilling the range request
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", subdirWithMixedBlockFiles.MustGetCidWithCodec(0x70, "subdir", "multiblock.txt")).
				Query("format", "car").
				Query("dag-scope", "entity").
				Query("entity-bytes", "512:*"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetDescendantsCids("subdir", "multiblock.txt")[2:])...,
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with entity-bytes equivalent to a HTTP Range Request for the middle of a file",
			Hint: `
				The response MUST contain only the minimal set of blocks necessary for fulfilling the range request
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", subdirWithMixedBlockFiles.MustGetCidWithCodec(0x70, "subdir", "multiblock.txt")).
				Query("format", "car").
				Query("dag-scope", "entity").
				Query("entity-bytes", "512:1023"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetDescendantsCids("subdir", "multiblock.txt")[2:4])...,
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with entity-bytes equivalent to a HTTP Range Request for the middle of a file (negative ending)",
			Hint: `
				The response MUST contain only the minimal set of blocks necessary for fulfilling the range request
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", subdirWithMixedBlockFiles.MustGetCidWithCodec(0x70, "subdir", "multiblock.txt")).
				Query("format", "car").
				Query("dag-scope", "entity").
				Query("entity-bytes", "512:-256"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetDescendantsCids("subdir", "multiblock.txt")[2:4])...,
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with entity-bytes equivalent to HTTP Suffix Range Request for part of a file",
			Hint: `
				The response MUST contain only the minimal set of blocks necessary for fulfilling the range request
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", subdirWithMixedBlockFiles.MustGetCidWithCodec(0x70, "subdir", "multiblock.txt")).
				Query("format", "car").
				Query("dag-scope", "entity").
				Query("entity-bytes", "-5:*"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetDescendantsCids("subdir", "multiblock.txt")[3:])...,
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with entity-bytes requesting a negative range bigger than the length of a file",
			Hint: `
				When range starts on negative index that makes it bigger than the file
				the request is truncated and starts at the beginning of a file.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", subdirWithMixedBlockFiles.MustGetCidWithCodec(0x70, "subdir", "multiblock.txt")).
				Query("format", "car").
				Query("dag-scope", "entity").
				Query("entity-bytes", "-9999:*"), // multiblock.txt size is 1026 (4*256+2)
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks( // expect entire file
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"), // dag-pb root of the file DAG
								"bafkreie5noke3mb7hqxukzcy73nl23k6lxszxi5w3dtmuwz62wnvkpsscm",    // 256 chunk
								"bafkreih4ephajybraj6wnxsbwjwa77fukurtpl7oj7t7pfq545duhot7cq",    // 256
								"bafkreigu7buvm3cfunb35766dn7tmqyh2um62zcio63en2btvxuybgcpue",    // 256
								"bafkreicll3huefkc3qnrzeony7zcfo7cr3nbx64hnxrqzsixpceg332fhe",    // 256
								"bafkreifst3pqztuvj57lycamoi7z34b4emf7gawxs74nwrc2c7jncmpaqm",    // 2
							)...).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with entity-bytes requesting a range from the end of a file that is bigger than a file itself",
			Hint: `
				The response MUST contain only the minimal set of blocks necessary for fulfilling the range request,
				everything before file start is ignored and the explicit end of the range is respected.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", subdirWithMixedBlockFiles.MustGetCidWithCodec(0x70, "subdir", "multiblock.txt")).
				Query("format", "car").
				Query("dag-scope", "entity").
				Query("entity-bytes", "-9999:-3"), // multiblock.txt size is 1026 (4*256+2)
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"), // dag-pb root of the file DAG
								"bafkreie5noke3mb7hqxukzcy73nl23k6lxszxi5w3dtmuwz62wnvkpsscm",    // 256 chunk
								"bafkreih4ephajybraj6wnxsbwjwa77fukurtpl7oj7t7pfq545duhot7cq",    // 256
								"bafkreigu7buvm3cfunb35766dn7tmqyh2um62zcio63en2btvxuybgcpue",    // 256
								"bafkreicll3huefkc3qnrzeony7zcfo7cr3nbx64hnxrqzsixpceg332fhe",    // 256
								// skip "bafkreifst3pqztuvj57lycamoi7z34b4emf7gawxs74nwrc2c7jncmpaqm",    // 2
							)...).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with entity-bytes requesting only the blocks for the first byte of a file",
			Hint: `
				The response MUST contain only the first block of the file.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", subdirWithMixedBlockFiles.MustGetCidWithCodec(0x70, "subdir", "multiblock.txt")).
				Query("format", "car").
				Query("dag-scope", "entity").
				Query("entity-bytes", "0:0"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetDescendantsCids("subdir", "multiblock.txt")[0])...,
						).
						Exactly().
						InThatOrder(),
				),
		},
	}

	RunWithSpecs(t, helpers.StandardCARTestTransforms(t, tests), specs.TrustlessGatewayCAR)
}

func TestTrustlessCarOrderAndDuplicates(t *testing.T) {
	tooling.LogTestGroup(t, GroupBlockCar)

	dirWithDuplicateFiles := car.MustOpenUnixfsCar("trustless_gateway_car/dir-with-duplicate-files.car")
	// This array is defined at the SPEC level and should not depend on library behavior
	// See the recipe for `dir-with-duplicate-files.car`
	multiblockCIDs := []string{
		"bafkreie5noke3mb7hqxukzcy73nl23k6lxszxi5w3dtmuwz62wnvkpsscm",
		"bafkreih4ephajybraj6wnxsbwjwa77fukurtpl7oj7t7pfq545duhot7cq",
		"bafkreigu7buvm3cfunb35766dn7tmqyh2um62zcio63en2btvxuybgcpue",
		"bafkreicll3huefkc3qnrzeony7zcfo7cr3nbx64hnxrqzsixpceg332fhe",
		"bafkreifst3pqztuvj57lycamoi7z34b4emf7gawxs74nwrc2c7jncmpaqm",
	}

	tests := SugarTests{
		{
			Name: "GET CAR with order=dfs and dups=y of UnixFS Directory With Duplicate Files",
			Hint: `
				The response MUST contain all the blocks found during traversal even if they
				are duplicate. In this test, a directory that contains duplicate files is
				requested. The blocks corresponding to the duplicate files must be returned.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", dirWithDuplicateFiles.MustGetCid()).
				Header("Accept", "application/vnd.ipld.car; version=1; order=dfs; dups=y"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").Contains("application/vnd.ipld.car"),
					Header("Content-Type").Contains("order=dfs"),
					Header("Content-Type").Contains("dups=y"),
				).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlock(dirWithDuplicateFiles.MustGetCid()).
						HasBlock(dirWithDuplicateFiles.MustGetCid("ascii.txt")). // ascii.txt = ascii-copy.txt
						HasBlock(dirWithDuplicateFiles.MustGetCid("ascii-copy.txt")).
						HasBlock(dirWithDuplicateFiles.MustGetCid("hello.txt")).
						HasBlock(dirWithDuplicateFiles.MustGetCid("multiblock.txt")).
						HasBlocks(multiblockCIDs...).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with order=dfs and dups=n of UnixFS Directory With Duplicate Files",
			Hint: `
				The response MUST NOT contain duplicate blocks. Tested
				directory contains duplicate files. The blocks corresponding to
				the duplicate files must be returned only ONCE.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", dirWithDuplicateFiles.MustGetCid()).
				Header("Accept", "application/vnd.ipld.car; version=1; order=dfs; dups=n"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").Contains("application/vnd.ipld.car"),
					Header("Content-Type").Contains("order=dfs"),
					Header("Content-Type").Contains("dups=n"),
				).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlock(dirWithDuplicateFiles.MustGetCid()).
						HasBlock(dirWithDuplicateFiles.MustGetCid("ascii.txt")). // ascii.txt = ascii-copy.txt
						HasBlock(dirWithDuplicateFiles.MustGetCid("hello.txt")).
						HasBlock(dirWithDuplicateFiles.MustGetCid("multiblock.txt")).
						HasBlocks(multiblockCIDs...).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR smoke-test with order=unk of UnixFS Directory",
			Hint: `
				The order=unk is usually used by gateway to explicitly indicate
				it does not guarantee any block order. In this case, we use it
				for basic smoke-test to confirm support of IPIP-412. The
				response for request with explicit order=unk MUST include an
				explicit order in returned Content-Type and contain all the
				blocks required to construct the requested CID. However, the
				gateway is free to return default ordering of own choosing,
				which means the returned blocks can be in any order and
				duplicates MAY occur.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", dirWithDuplicateFiles.MustGetCid()).
				Header("Accept", "application/vnd.ipld.car; version=1; order=unk"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").Contains("application/vnd.ipld.car"),
					Header("Content-Type").Contains("order="),
				).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlock(dirWithDuplicateFiles.MustGetCid()).
						HasBlock(dirWithDuplicateFiles.MustGetCid("ascii.txt")).
						HasBlock(dirWithDuplicateFiles.MustGetCid("ascii-copy.txt")).
						HasBlock(dirWithDuplicateFiles.MustGetCid("hello.txt")).
						HasBlock(dirWithDuplicateFiles.MustGetCid("multiblock.txt")).
						HasBlocks(multiblockCIDs...),
				),
		},
		{
			Name: "GET CAR with order=dfs and dups=y of identity CID",
			Hint: `
				Identity hashes MUST never be manifested as read blocks.
				These are virtual ones and even when dups=y is set, they never
				should be returned in CAR response body.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", "bafkqaf3imvwgy3zaneqgc3janfxgy2lomvscay3jmqfa").
				Header("Accept", "application/vnd.ipld.car; dups=y"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").Contains("application/vnd.ipld.car"),
					Header("Content-Type").Contains("dups=y"),
				).
				Body(
					IsCar().
						IgnoreRoots().
						Exactly().
						InThatOrder(),
				),
		},
		// Tests for car-order and car-dups URL query parameters (IPIP-0523)
		{
			Name: "GET CAR with ?format=car respects Accept header order and dups params",
			Hint: `
				When format=car is used, the Accept header can still provide CAR-specific
				parameters like order and dups. The response MUST contain all the blocks
				found during traversal even if they are duplicate.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}?format=car", dirWithDuplicateFiles.MustGetCid()).
				Header("Accept", "application/vnd.ipld.car; version=1; order=dfs; dups=y"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").Contains("application/vnd.ipld.car"),
					Header("Content-Type").Contains("order=dfs"),
					Header("Content-Type").Contains("dups=y"),
				).
				Body(
					IsCar().
						IgnoreRoots().
						HasBlock(dirWithDuplicateFiles.MustGetCid()).
						HasBlock(dirWithDuplicateFiles.MustGetCid("ascii.txt")).
						HasBlock(dirWithDuplicateFiles.MustGetCid("ascii-copy.txt")).
						HasBlock(dirWithDuplicateFiles.MustGetCid("hello.txt")).
						HasBlock(dirWithDuplicateFiles.MustGetCid("multiblock.txt")).
						HasBlocks(multiblockCIDs...).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with ?car-order=dfs takes precedence over order=unk in Accept",
			Spec: "https://specs.ipfs.tech/http-gateways/trustless-gateway/#car-order-request-query-parameter",
			Hint: `
				Per IPIP-0523, URL query parameters should take precedence over Accept header parameters.
				When car-order=dfs is in URL and order=unk is in Accept, response should have order=dfs.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}?format=car&car-order=dfs", dirWithDuplicateFiles.MustGetCid()).
				Header("Accept", "application/vnd.ipld.car; order=unk"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").Contains("application/vnd.ipld.car"),
					Header("Content-Type").Contains("order=dfs"),
				),
		},
		{
			Name: "GET CAR with ?car-dups=y takes precedence over dups=n in Accept",
			Spec: "https://specs.ipfs.tech/http-gateways/trustless-gateway/#car-dups-request-query-parameter",
			Hint: `
				Per IPIP-0523, URL query parameters should take precedence over Accept header parameters.
				When car-dups=y is in URL and dups=n is in Accept, response should have dups=y.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}?format=car&car-dups=y", dirWithDuplicateFiles.MustGetCid()).
				Header("Accept", "application/vnd.ipld.car; dups=n"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").Contains("application/vnd.ipld.car"),
					Header("Content-Type").Contains("dups=y"),
				),
		},
	}

	RunWithSpecs(t, tests, specs.TrustlessGatewayCAROptional)
}

func TestTrustlessCarFormatPrecedence(t *testing.T) {
	tooling.LogTestGroup(t, GroupBlockCar)

	fixture := car.MustOpenUnixfsCar("gateway-raw-block.car")

	tests := SugarTests{
		{
			Name: "GET with format=car query parameter takes precedence over Accept header",
			Spec: "https://specs.ipfs.tech/http-gateways/trustless-gateway/#format-request-query-parameter",
			Hint: `
			Per IPIP-0523, the format query parameter should be preferred over the
			Accept header when both are present. This test verifies that format=car
			overrides Accept: application/vnd.ipld.raw.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}?format=car", fixture.MustGetCid("dir")).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
				),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").
						Contains("application/vnd.ipld.car"),
				),
		},
		{
			Name: "GET with format=raw query parameter takes precedence over Accept header",
			Spec: "https://specs.ipfs.tech/http-gateways/trustless-gateway/#format-request-query-parameter",
			Hint: `
			Per IPIP-0523, the format query parameter should be preferred over the
			Accept header when both are present. This test verifies that format=raw
			overrides Accept: application/vnd.ipld.car.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}?format=raw", fixture.MustGetCid("dir")).
				Headers(
					Header("Accept", "application/vnd.ipld.car"),
				),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").
						Equals("application/vnd.ipld.raw"),
				),
		},
	}

	RunWithSpecs(t, tests, specs.TrustlessGatewayCAR)
}

// TODO: this feels like it could be an internal detail of HasBlocks
func flattenStrings(t *testing.T, values ...interface{}) []string {
	var res []string
	for _, v := range values {
		switch tv := v.(type) {
		case string:
			res = append(res, tv)
		case []string:
			res = append(res, tv...)
		default:
			t.Fatal("only strings and string slices supported, this should be unreachable")
		}
	}
	return res
}
