// https://github.com/ipfs/kubo/blob/master/test/sharness/t0116-gateway-cache.sh
import { run, TestRequestSuiteDefinition } from "declarative-e2e-test";
import { Fixture } from "../util/fixtures.js";
import { config } from "./config.js";

// TODO: this pattern with wrap directory is confusing.
const root = Fixture.get("root2");
const root2 = root.get("root2");
const root3 = root.get("root2/root3");
const root4 = root.get("root2/root3/root4");
const index = root.get("root2/root3/root4/index.html");

const R = (x: string) => new RegExp(`^${x}$`);

const test: TestRequestSuiteDefinition = {
  "GET /ipfs/": {
    tests: {
      unixfs: {
        tests: {
          "GET for /ipfs/ unifx dir listing succeeds": {
            url: `/ipfs/${root.cid}/root2/root3/`,
            expect: [
              200,
              {
                headers: {
                  "X-Ipfs-Path": `/ipfs/${root.cid}/root2/root3/`,
                  "X-Ipfs-Roots": `${root.cid},${root2.cid},${root3.cid}`,
                  Etag: R(`"DirIndex-.+_CID-${root3.cid}"`),
                  // TODO: https://github.com/ipfs/kubo/blob/799e5ac0a5a6600e844aad585282ad23789a88e7/test/sharness/t0116-gateway-cache.sh#L87
                  // "Cache-Control": "public, max-age=TBD",
                },
              },
            ],
          },
          "GET for /ipfs/ unixfs dir with index.html succeeds": {
            url: `/ipfs/${root.cid}/root2/root3/root4/`,
            expect: [
              200,
              {
                headers: {
                  "X-Ipfs-Path": `/ipfs/${root.cid}/root2/root3/root4/`,
                  "X-Ipfs-Roots": `${root.cid},${root2.cid},${root3.cid},${root4.cid}`,
                  "Cache-Control": "public, max-age=29030400, immutable",
                  Etag: `"${root4.cid}"`,
                },
              },
            ],
          },
          "GET for /ipfs/ unixfs file succeeds": {
            url: `/ipfs/${root.cid}/root2/root3/root4/index.html`,
            expect: [
              200,
              {
                headers: {
                  "X-Ipfs-Path": `/ipfs/${root.cid}/root2/root3/root4/index.html`,
                  "X-Ipfs-Roots": `${root.cid},${root2.cid},${root3.cid},${root4.cid},${index.cid}`,
                  "Cache-Control": "public, max-age=29030400, immutable",
                  Etag: `"${index.cid}"`,
                },
              },
            ],
          },
        },
      },
    },
  },
  "If-None-Match (return 304 Not Modified when client sends matching Etag they already have)":
    {
      tests: {
        "GET for /ipfs/ file with matching Etag in If-None-Match returns 304 Not Modified":
          {
            url: `/ipfs/${root.cid}/root2/root3/root4/index.html`,
            headers: {
              "If-None-Match": `"${index.cid}"`,
            },
            expect: [304],
          },
        "GET for /ipfs/ dir with index.html file with matching Etag in If-None-Match returns 304 Not Modified":
          {
            url: `/ipfs/${root.cid}/root2/root3/root4/`,
            headers: {
              "If-None-Match": `"${root4.cid}"`,
            },
            expect: [304],
          },
        "GET for /ipfs/ file with matching third Etag in If-None-Match returns 304 Not Modified":
          {
            url: `/ipfs/${root.cid}/root2/root3/root4/index.html`,
            headers: {
              "If-None-Match": `"fakeEtag1", "fakeEtag2", "${index.cid}"`,
            },
            expect: [304],
          },
        "GET for /ipfs/ file with matching weak Etag in If-None-Match returns 304 Not Modified":
          {
            url: `/ipfs/${root.cid}/root2/root3/root4/index.html`,
            headers: {
              "If-None-Match": `W/"${index.cid}"`,
            },
            expect: [304],
          },
        "GET for /ipfs/ file with wildcard Etag in If-None-Match returns 304 Not Modified":
          {
            url: `/ipfs/${root.cid}/root2/root3/root4/index.html`,
            headers: {
              "If-None-Match": `*`,
            },
            expect: [304],
          },
        "GET for /ipfs/ dir listing with matching weak Etag in If-None-Match returns 304 Not Modified":
          {
            url: `/ipfs/${root.cid}/root2/root3/`,
            headers: {
              "If-None-Match": `W/"${root3.cid}"`,
            },
            expect: [304],
          },
      },
    },
};

console.log("Running test: cache.spec.ts");
run(test, config);
