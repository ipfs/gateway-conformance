import { ok } from "assert";
import { run, TestRequestSuiteDefinition } from "declarative-e2e-test";
import { Response } from "supertest";
import { config } from "./config.js";
import fixture, { blockAsString, blockSize } from "./fixtures.js";

const IPLD_RAW_TYPE = "application/vnd.ipld.raw";

const test: TestRequestSuiteDefinition = {
  "Test HTTP Gateway Raw Block (application/vnd.ipld.raw) Support": {
    tests: {
      "GET with format=raw param returns a raw block": {
        url: `/ipfs/${fixture.root._cid}/dir?format=raw`,
        expect: [200, blockAsString(fixture.root.dir)],
      },
      "GET for application/vnd.ipld.raw returns a raw block": {
        url: `/ipfs/${fixture.root._cid}/dir`,
        headers: { accept: IPLD_RAW_TYPE },
        expect: [200, blockAsString(fixture.root.dir)],
      },
      "GET response for application/vnd.ipld.raw has expected response headers":
        {
          url: `/ipfs/${fixture.root._cid}/dir/ascii.txt`,
          headers: { accept: IPLD_RAW_TYPE },
          expect: [
            200,
            {
              headers: {
                "content-type": IPLD_RAW_TYPE,
                "content-length": blockSize(
                  fixture.root.dir["ascii.txt"]
                ).toString(),
                "content-disposition": new RegExp(
                  `attachment;\\s*filename="${fixture.root.dir["ascii.txt"]._cid}\\.bin`
                ),
                "x-content-type-options": "nosniff",
              },
              body: blockAsString(fixture.root.dir["ascii.txt"]),
            },
          ],
        },
      "GET for application/vnd.ipld.raw with query filename includes Content-Disposition with custom filename":
        {
          url: `/ipfs/${fixture.root._cid}/dir/ascii.txt?filename=foobar.bin`,
          headers: { accept: IPLD_RAW_TYPE },
          expect: [
            200,
            {
              headers: {
                "content-disposition": new RegExp(
                  `attachment;\\s*filename="foobar\\.bin"`
                ),
              },
            },
          ],
        },
      "GET response for application/vnd.ipld.raw has expected caching headers":
        {
          url: `/ipfs/${fixture.root._cid}/dir/ascii.txt`,
          headers: { accept: IPLD_RAW_TYPE },
          expect: [
            200,
            {
              headers: {
                etag: `"${fixture.root.dir["ascii.txt"]._cid}.raw"`,
                "x-ipfs-path": `/ipfs/${fixture.root._cid}/dir/ascii.txt`,
                "x-ipfs-roots": new RegExp(fixture.root._cid),
              },
            },
            function (response: Response) {
              const cachePragmata = (
                response.headers["cache-control"] || ""
              ).split(/\s*,\s*/);
              Object.entries({
                "public pragma": (str: string) =>
                  str.toLowerCase() === "public",
                "immutable pragma": (str: string) =>
                  str.toLowerCase() === "immutable",
                "max-age pragma": (str: string) => {
                  if (!/^max-age=/i.test(str)) return false;
                  if (parseInt(str.replace("max-age=", ""), 10) < 29030400)
                    return false; // at least 29030400
                  return true;
                },
              }).forEach(([label, func]) =>
                ok(cachePragmata.find(func), label)
              );
            },
          ],
        },
    },
  },
};

console.log("Running test: raw-block.spec.ts");
run(test, config);
