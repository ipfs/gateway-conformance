// https://github.com/ipfs/kubo/blob/master/test/sharness/t0118-gateway-car.sh
import { run, TestRequestSuiteDefinition } from "declarative-e2e-test";
import { config } from "./config.js";
import fixture, { dagAsString } from "./fixtures.js";

const IPLD_CAR_TYPE = "application/vnd.ipld.car";

// TODO:
// Take into account this comment and how it might impact our code.
//
// CAR stream is not deterministic, as blocks can arrive in random order,
// but if we have a small file that fits into a single block, and export its CID
// we will get a CAR that is a deterministic array of bytes.

const test: TestRequestSuiteDefinition = {
  "Test HTTP Gateway CAR (application/vnd.ipld.car) Support": {
    tests: {
      "GET a reference DAG with dag-cbor+dag-pb+raw blocks as CAR": {
        tests: {
          // This test uses official CARv1 fixture from https://ipld.io/specs/transport/car/fixture/carv1-basic/
          // TODO
        },
      },
      "GET unixfs file as CAR (by using a single file we ensure deterministic result that can be compared byte-for-byte)":
        {
          tests: {
            "GET with format=car param returns a CARv1 stream": {
              url: `/ipfs/${fixture.car._cid}/subdir/ascii.txt?format=car`,
              expect: [200, dagAsString(fixture.car.subdir["ascii.txt"])],
            },
            "GET for application/vnd.ipld.car returns a CARv1 stream": {
              url: `/ipfs/${fixture.car._cid}/subdir/ascii.txt`,
              headers: { accept: IPLD_CAR_TYPE },
              expect: [200, dagAsString(fixture.car.subdir["ascii.txt"])],
            },
            "GET for application/vnd.ipld.raw version=1 returns a CARv1 stream":
              {
                url: `/ipfs/${fixture.car._cid}/subdir/ascii.txt`,
                headers: { accept: `${IPLD_CAR_TYPE};version=1` },
                expect: [200, dagAsString(fixture.car.subdir["ascii.txt"])],
              },
            "GET for application/vnd.ipld.raw version=1 returns a CARv1 stream (with whitespace)":
              {
                url: `/ipfs/${fixture.car._cid}/subdir/ascii.txt`,
                headers: { accept: `${IPLD_CAR_TYPE}; version=1` },
                expect: [200, dagAsString(fixture.car.subdir["ascii.txt"])],
              },
            "GET for application/vnd.ipld.raw version=2 returns HTTP 400 Bad Request error":
              {
                url: `/ipfs/${fixture.car._cid}/subdir/ascii.txt`,
                headers: { accept: `${IPLD_CAR_TYPE};version=2` },
                expect: [400, /unsupported CAR version/],
              },
          },
        },
      "GET unixfs directory as a CAR with DAG and some selector": {
        tests: {
          // TODO: this is basic test for "full" selector, we will add support for custom ones in https://github.com/ipfs/go-ipfs/issues/8769
          "GET for application/vnd.ipld.car with unixfs dir returns a CARv1 stream with full DAG":
            {
              url: `/ipfs/${fixture.car._cid}`,
              headers: { accept: IPLD_CAR_TYPE },
              expect: [200, dagAsString(fixture.car)],
            },
        },
      },
      "Make sure expected HTTP headers are returned with the CAR bytes": {
        url: `/ipfs/${fixture.car._cid}/subdir/ascii.txt`,
        headers: { accept: IPLD_CAR_TYPE },
        expect: {
          headers: {
            "Content-Type": {
              info: "GET response for application/vnd.ipld.car has expected Content-Type",
              value: `${IPLD_CAR_TYPE}; version=1`,
            },
            "Content-Length": {
              info: "GET response for application/vnd.ipld.car includes no Content-Length. CAR is streamed, gateway may not have the entire thing, unable to calculate total size.",
              value: null,
            },
            "Content-Disposition": {
              info: "GET response for application/vnd.ipld.car includes Content-Disposition",
              value: new RegExp(
                `attachment; filename=\"${fixture.car.subdir["ascii.txt"]._cid}.car\"`
              ),
            },
            "X-Content-Type-Options": {
              info: "GET response for application/vnd.ipld.car includes nosniff hint",
              value: "nosniff",
            },
            "Accept-Ranges": {
              info: `GET response for application/vnd.ipld.car includes Accept-Ranges header.
              CAR is streamed, gateway may not have the entire thing, unable to support range-requests
              Partial downloads and resumes should be handled using IPLD selectors: https://github.com/ipfs/go-ipfs/issues/8769`,
              value: "none",
            },
            ETag: {
              info: "GET response for application/vnd.ipld.car includes a weak Etag",
              value: new RegExp(
                `W/"${fixture.car.subdir["ascii.txt"]._cid}.car"`
              ),
            },
            "X-Ipfs-Path": {
              info: "GET response for application/vnd.ipld.car includes X-Ipfs-Path and X-Ipfs-Roots. (basic checks, detailed behavior for some fields is tested in  t0116-gateway-cache.sh)",
              value: /.+/,
            },
            "X-Ipfs-Roots": {
              info: "GET response for application/vnd.ipld.car includes X-Ipfs-Path and X-Ipfs-Roots. (basic checks, detailed behavior for some fields is tested in  t0116-gateway-cache.sh)",
              value: /.+/,
            },
            "Cache-Control": {
              info: "GET response for application/vnd.ipld.car includes same Cache-Control as a block or a file",
              value: "public, max-age=29030400, immutable",
            },
          },
        },
      },
      "GET for application/vnd.ipld.car with query filename includes Content-Disposition with custom filename":
        {
          url: `/ipfs/${fixture.car._cid}/subdir/ascii.txt?filename=foobar.car`,
          headers: { accept: IPLD_CAR_TYPE },
          expect: {
            headers: {
              "Content-Disposition": new RegExp(
                `attachment; filename=\"foobar.car\"`
              ),
            },
          },
        },
    },
  },
};

console.log("Running test: raw-block.spec.ts");
run(test, config);
