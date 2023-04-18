# Workshop IPFS Thing 2023

- [Workshop IPFS Thing 2023](#workshop-ipfs-thing-2023)
  - [Intro to the gateway test suite](#intro-to-the-gateway-test-suite)
    - [Current State](#current-state)
    - [Design](#design)
      - [Simple API \& Mostly Data](#simple-api--mostly-data)
      - [Emphasis on detailed reporting](#emphasis-on-detailed-reporting)
      - [Applies to "any" type of gateway](#applies-to-any-type-of-gateway)
      - [Relation with the Specs](#relation-with-the-specs)
  - [Write our first 3 tests (guided walkthrough)](#write-our-first-3-tests-guided-walkthrough)
    - [Green Run](#green-run)
      - [Fix our first test](#fix-our-first-test)
      - [Implement our second test](#implement-our-second-test)
      - [Implement a more complex test](#implement-a-more-complex-test)
  - [Write a test that relies on subdomain or dnslink](#write-a-test-that-relies-on-subdomain-or-dnslink)
    - [Setup the env](#setup-the-env)
    - [Implement a subdomain test](#implement-a-subdomain-test)
  - [Write a new spec test](#write-a-new-spec-test)
    - [TODOs](#todos)

## Intro to the gateway test suite

`gateway-conformance` is a tool designed to test if an IPFS Gateway implementation complies with the IPFS Gateway Specification correctly. The tool is distributed as a Docker image and a GitHub Action(s).

<https://github.com/ipfs/gateway-conformance>

### Current State

Many Kubo Sharness tests have been ported to go.

The suite is used in CI:

- [kubo](https://github.com/ipfs/kubo/actions/workflows/gateway-conformance.yml)
- [bifrost-gateway](https://github.com/ipfs/bifrost-gateway/actions/workflows/gateway-conformance.yml)
- [boxo](https://github.com/ipfs/boxo/actions/workflows/gateway-conformance.yml)

### Design

#### Simple API & Mostly Data

- Make sure it's easy to contribute,
- Make sure it's easy to transform tests when needed (generate more test cases from simple definitions, upgrade or change the APIs easily, for example, adding interop with specs).

#### Emphasis on detailed reporting

Currently, error dumps contain quick, actionable feedback on errors,
and we output detailed markdown of test passing/failings.

Later have metrics about how "conforming" a gateway is.

#### Applies to "any" type of gateway

Enable / Disable specs like subdomain, DNS links, etc.

Configurable domain URL: We can test a domain gateway running on local env (<http://127.0.0.1>) and can run the same test suite on a live gateway as well (<http://dweb.link>)

TODO: prove this; we don't at the moment.

#### Relation with the Specs

We've been porting the Kubo test suite to make them reusable across implementations and make them easier to scale.

The next step will be to interact directly with the specs:

- we have hints of this: the redirect test suite relies on the car file provided in the specs.
- eventually, we want a way to link a "phrase" in the spec to a test in the test suite.
  - Contribution welcome.

## Write our first 3 tests (guided walkthrough)

If you see errors when generating the `./fixtures.car`: remove the file and re-run

### Green Run

Start a Kubo node for local dev and provision the node.
We'll use the makefile:

`make test-kubo`

This will:

- build the CLI
- provision the Kubo gateway: import all the fixtures on your local daemon
- run the test suite (without subdomains tests)

You should see an error that we're going to fix!

#### Fix our first test

> Introduce the reporting and the test format

We started porting a test from kubo sharness (115 - gateway dir listing)

Original shell script: `t0115-gateway-dir-listing.sh`
We moved the fixture in the gateway conformance repo, created the test file, and prepared a few tests for you. It lives in `t0115_gateway_dir_listing_test.go`.

We started porting the first test:

```bash
test_expect_success "path gw: backlink on root CID should be hidden" '
curl -sD - http://127.0.0.1:$GWAY_PORT/ipfs/${DIR_CID}/ > list_response &&
test_should_contain "Index of" list_response &&
test_should_not_contain "<a href=\"/ipfs/$DIR_CID/\">..</a>" list_response
'
```

Which looks like this:

```go
{
 Name: "path gw: backlink on root CID should be hidden",
 Hint: `
 this test is written for the workshop; it will fail by default.
 But we can use it to show a rough idea of how to write tests.
 `,
 Request: Request().
  Path("/ipfs/%s/", dir.Cid()),
 Response: Expect().
  Status(202).
  Body(
   And(
    Contains("Index of"),
    Not(Contains("<a href=\"/ipfs/%s/\">..</a>", dir.Cid())),
   ),
  ),
},
```

But there is an error in the test. You should see this in your error logs:

```
--- FAIL: TestGatewayDirListingOnPathGateway (0.02s)
    --- FAIL: TestGatewayDirListingOnPathGateway/path_gw:_backlink_on_root_CID_should_be_hidden (0.02s)
        /Users/laurent/dev/plabs/gateway-conformance/tests/report.go:89:
            Name: path gw: backlink on root CID should be hidden
            Hint:

            Error: Status code is not 202. It is 200

            Request:
            {
              "method": "GET",
              "path": "/ipfs/bafybeig6ka5mlwkl4subqhaiatalkcleo4jgnr3hqwvpmsqfca27cijp3i/"
            }

            Expected Response:
            {
              "statusCode": 202,
              "body": {
                "Checks": [
                  {
                    "Value": "Index of"
                  },
                  {}
                ]
              }
            }

            Actual Request:
            GET //ipfs/bafybeig6ka5mlwkl4subqhaiatalkcleo4jgnr3hqwvpmsqfca27cijp3i/ HTTP/1.1
            Host: 127.0.0.1:8080
            User-Agent: Go-http-client/1.1
            Accept-Encoding: gzip



            Actual Response:
            HTTP/1.1 200 OK
            Transfer-Encoding: chunked
```

Fix the test and add a new one.

#### Implement our second test

> Start from scratch and write a new test

No move on to the next test, just below:

```
test_expect_success "path gw: redirect dir listing to URL with trailing slash" '
curl -sD - http://127.0.0.1:$GWAY_PORT/ipfs/${DIR_CID}/ą/ę > list_response &&
test_should_contain "HTTP/1.1 301 Moved Permanently" list_response &&
test_should_contain "Location: /ipfs/${DIR_CID}/%c4%85/%c4%99/" list_response
'
```

It should be easy to port.

#### Implement a more complex test

> Introduce Headers, Body, etc.

These tests are pretty simple. We've implemented the hard part first!

Open `t0117_gateway_block_test.go`. There is a `// TODO` you can follow along.

## Write a test that relies on a subdomain or DNS link

These require extra configuration and rely on the specs.

### Setup the env

:warning: the following command will change your ipfs configuration.

Run the script `./kubo-config.example.sh`, which will update your configuration
and print an env variable you may use for dnslinking.

Restart the Kubo daemon with this env. It should look something like this:

```
IPFS_NS_MAP=dnslink-enabled-on-fqdn.example.com:/ipfs/QmYBhLYDwVFvxos9h8CGU2ibaY66QNgv8hpfewxaQrPiZj ipfs daemon
```

Then `make test-kubo-subdomains` will run the test with subdomain specs enabled.

// TODO Add the subdomain test to the gateway dir listing. Show how the sugar that generates more tests.

### Implement a subdomain test

Constraint: make this configurable; we should be able to use

- example.com and localhost for local dev,
- but also dweb.link, cloudflare-ipfs.com, etc.

Solution: construct an URL, then use proxying, host tweaks, etc.

test implementer: construct URLs (we can't guess these)
use `helpers.UnwrapSubdomainTests` to generate more tests (thanks to the data-driven approach, we can compose, etc.).

## Write a new spec test

https://specs.ipfs.tech/http-gateways/

What about testing https://specs.ipfs.tech/http-gateways/trustless-gateway/ ?


### TODOs

When we run the test suite, it fails. What advice can we share to fix this? I (Laurent) run the tests in an IDE. It's easy to find there. If you run in CLI, how do you fix it?
