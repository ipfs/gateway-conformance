package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/helpers"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestTrustlessCarPathing(t *testing.T) {
	subdirTwoSingleBlockFilesFixture := car.MustOpenUnixfsCar("t0118/subdir-with-two-single-block-files.car")
	singleLayerHamtMultiBlockFilesFixture := car.MustOpenUnixfsCar("t0118/single-layer-hamt-with-multi-block-files.car")
	dirWithDagCborWithLinksFixture := car.MustOpenUnixfsCar("t0118/dir-with-dag-cbor-with-links.car")

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
						HasRoot(subdirTwoSingleBlockFilesFixture.MustGetCid("subdir", "ascii.txt")).
						MightHaveNoRoots().
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
						HasRoot(singleLayerHamtMultiBlockFilesFixture.MustGetCid("685.txt")).
						MightHaveNoRoots().
						HasBlocks(flattenStrings(t,
							singleLayerHamtMultiBlockFilesFixture.MustGetCid(),
							singleLayerHamtMultiBlockFilesFixture.MustGetCIDsInHAMTTraversal(nil, "685.txt"),
							singleLayerHamtMultiBlockFilesFixture.MustGetCid("685.txt"),
							singleLayerHamtMultiBlockFilesFixture.MustGetChildrenCids("685.txt"))...,
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
						HasRoot(dirWithDagCborWithLinksFixture.MustGetCid("document", "files", "single")).
						MightHaveNoRoots().
						HasBlocks(flattenStrings(t,
							dirWithDagCborWithLinksFixture.MustGetCid("document"),
							dirWithDagCborWithLinksFixture.MustGetCid("document", "files", "single"),
						)...).
						Exactly().
						InThatOrder(),
				),
		},
	}

	RunWithSpecs(t, helpers.StandardCARTestTransforms(t, tests), specs.TrustlessGatewayCAR)
}

func TestTrustlessCarDagScopeBlock(t *testing.T) {
	subdirTwoSingleBlockFilesFixture := car.MustOpenUnixfsCar("t0118/subdir-with-two-single-block-files.car")
	singleLayerHamtMultiBlockFilesFixture := car.MustOpenUnixfsCar("t0118/single-layer-hamt-with-multi-block-files.car")

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
						HasRoot(subdirTwoSingleBlockFilesFixture.MustGetCid("subdir")).
						MightHaveNoRoots().
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
						HasRoot(subdirTwoSingleBlockFilesFixture.MustGetCid("subdir", "ascii.txt")).
						MightHaveNoRoots().
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
						HasRoot(singleLayerHamtMultiBlockFilesFixture.MustGetCid("1.txt")).
						MightHaveNoRoots().
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
	subdirTwoSingleBlockFilesFixture := car.MustOpenUnixfsCar("t0118/subdir-with-two-single-block-files.car")
	singleLayerHamtMultiBlockFilesFixture := car.MustOpenUnixfsCar("t0118/single-layer-hamt-with-multi-block-files.car")
	subdirWithMixedBlockFiles := car.MustOpenUnixfsCar("t0118/subdir-with-mixed-block-files.car")
	dirWithDagCborWithLinksFixture := car.MustOpenUnixfsCar("t0118/dir-with-dag-cbor-with-links.car")

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
						HasRoot(subdirTwoSingleBlockFilesFixture.MustGetCid()).
						MightHaveNoRoots().
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
						HasRoot(singleLayerHamtMultiBlockFilesFixture.MustGetCid()).
						MightHaveNoRoots().
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
						HasRoot(subdirWithMixedBlockFiles.MustGetCid("subdir", "ascii.txt")).
						MightHaveNoRoots().
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
						HasRoot(subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt")).
						MightHaveNoRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid(),
								subdirWithMixedBlockFiles.MustGetCid("subdir"),
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetChildrenCids("subdir", "multiblock.txt"),
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
						HasRoot(dirWithDagCborWithLinksFixture.MustGetCid("document")).
						MightHaveNoRoots().
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
	subdirWithMixedBlockFiles := car.MustOpenUnixfsCar("t0118/subdir-with-mixed-block-files.car")

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
						HasRoot(subdirWithMixedBlockFiles.MustGetCid("subdir")).
						MightHaveNoRoots().
						HasBlocks(
							flattenStrings(
								t,
								subdirWithMixedBlockFiles.MustGetCid(),
								subdirWithMixedBlockFiles.MustGetCid("subdir"),
								subdirWithMixedBlockFiles.MustGetChildrenCids("subdir"),
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
						HasRoot(subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt")).
						MightHaveNoRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid(),
								subdirWithMixedBlockFiles.MustGetCid("subdir"),
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetChildrenCids("subdir", "multiblock.txt"),
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
	singleLayerHamtMultiBlockFilesFixture := car.MustOpenUnixfsCar("t0118/single-layer-hamt-with-multi-block-files.car")
	subdirWithMixedBlockFiles := car.MustOpenUnixfsCar("t0118/subdir-with-mixed-block-files.car")

	tests := SugarTests{
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
						HasRoot(subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt")).
						MightHaveNoRoots().
						HasBlocks(flattenStrings(t,
							subdirWithMixedBlockFiles.MustGetCid(),
							subdirWithMixedBlockFiles.MustGetCid("subdir"),
							subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
							subdirWithMixedBlockFiles.MustGetChildrenCids("subdir", "multiblock.txt"),
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
						HasRoot(singleLayerHamtMultiBlockFilesFixture.MustGetCid()).
						MightHaveNoRoots().
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
						HasRoot(subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt")).
						MightHaveNoRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetChildrenCids("subdir", "multiblock.txt")[2:])...,
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
						HasRoot(subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt")).
						MightHaveNoRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetChildrenCids("subdir", "multiblock.txt")[2:4])...,
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
						HasRoot(subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt")).
						MightHaveNoRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetChildrenCids("subdir", "multiblock.txt")[2:4])...,
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
						HasRoot(subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt")).
						MightHaveNoRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetChildrenCids("subdir", "multiblock.txt")[3:])...,
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with entity-bytes requesting a range from the end of a file",
			Hint: `
				The response MUST contain only the minimal set of blocks necessary for fulfilling the range request
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", subdirWithMixedBlockFiles.MustGetCidWithCodec(0x70, "subdir", "multiblock.txt")).
				Query("format", "car").
				Query("dag-scope", "entity").
				Query("entity-bytes", "-999999:-3"),
			Response: Expect().
				Status(200).
				Body(
					IsCar().
						HasRoot(subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt")).
						MightHaveNoRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetChildrenCids("subdir", "multiblock.txt")[:5])...,
						).
						Exactly().
						InThatOrder(),
				),
		},
		{
			Name: "GET CAR with entity-bytes requesting the first byte of a file",
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
						HasRoot(subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt")).
						MightHaveNoRoots().
						HasBlocks(
							flattenStrings(t,
								subdirWithMixedBlockFiles.MustGetCid("subdir", "multiblock.txt"),
								subdirWithMixedBlockFiles.MustGetChildrenCids("subdir", "multiblock.txt")[0])...,
						).
						Exactly().
						InThatOrder(),
				),
		},
	}

	RunWithSpecs(t, helpers.StandardCARTestTransforms(t, tests), specs.TrustlessGatewayCAR)
}

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
