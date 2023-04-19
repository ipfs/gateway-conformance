# Workshop IPFS Thing 2023

## Intro to the gateway test suite

`gateway-conformance` is a tool designed to test if an IPFS Gateway implementation complies with the IPFS Gateway Specification correctly. The tool is distributed as a Docker image, as well as a GitHub Action(s).

<https://github.com/ipfs/gateway-conformance>

### Current State

Many Kubo Sharness tests have been ported to go

The suite is used in CI:

- [kubo](https://github.com/ipfs/kubo/actions/workflows/gateway-conformance.yml)
- [bifrost-gateway](https://github.com/ipfs/bifrost-gateway/actions/workflows/gateway-conformance.yml)
- [boxo](https://github.com/ipfs/boxo/actions/workflows/gateway-conformance.yml)

### Design

#### Simple API & Mostly Data

- Make sure it's easy to contribute,
- Make sure it's easy to transform tests when needed (generate more test case from simple definitions, upgrade or change the APIs easily, for example adding interop with specs).

#### Emphasis on detailed reporting

Currently error dumps contains quick actionnable feedback on errors,
and we output detailed markdown of test passing / failings.

Later have metrics about how "conforming" a gateway is.

#### Applies to "any" type of gateway

Enable / Disable specs like subdomain, dnslinks, etc.

Configurable domain URL: We can test a domain gateway runnig on local env, (<http://127.0.0.1>), and are able run the same test suite on a live gateway aswell (<http://dweb.link>)

TODO: prove this, we don't at the moment.

#### Relation with the Specs

We've been porting kubo test suite to make them reusable accross implementations, and make them easier to scale.

Next step will be to interact directly with the specs:

- we have hints of this: the redirect test suite relies on car file provided in the specs.
- eventually we want a way to link a "phrase" in the spec to a test in the test suite.
  - Contribution welcome.

## Write our first 3 tests (guided walkthrough)

### Prepare the env

Prerequisites:
- gateway-conformance clone
- docker or Go 1.20
- kubo

#### Clone the repository

```bash
git clone git@github.com:ipfs/gateway-conformance.git
cd gateway-conformance
git checkout workshop
```

#### Start Kubo Gateway

```bash
ipfs daemon --offline &
```

#### Provision Kubo Gateway (import CAR fixtures)

```bash
make provision-kubo
```

#### Build the Gateway Conformance CLI

##### Docker (recommended)

```bash
make docker
```

##### Go 1.20 (alternative)

```bash
make gateway-conformance
```

#### Run the tests against Kubo Gateway

- pass the Gateway URL
- disable subdomain-gateway spec tests because we didn't configure Kubo to run on a subdomain

##### Docker (OSX)

```bash
./gc --gateway-url http://host.docker.internal:8080 --specs -subdomain-gateway
```

##### Docker (Linux)

```bash
./gc --gateway-url http://127.0.0.1:8080 --specs -subdomain-gateway
```

##### Go 1.20

```bash
./gateway-conformance --gateway-url http://127.0.0.1:8080 --specs -subdomain-gateway
```

### Green Run

On the first run, you should see an error. There is a typo in the test, we'll fix it now.

#### Fixing a test

> introducing reporting and the test format

We started porting a test from kubo sharness (115 - gateway dir listing, original shell script: `t0115-gateway-dir-listing.sh`).

We moved a static fixture to the gateway conformance repo, created a test file, and prepared a few test cases for you.

The file we'll be looking at lives at `tests/t0115_gateway_dir_listing_test.go`.

Here is how the test looked like in sharness:
```bash
test_expect_success "path gw: backlink on root CID should be hidden" '
curl -sD - http://127.0.0.1:$GWAY_PORT/ipfs/${DIR_CID}/ > list_response &&
test_should_contain "Index of" list_response &&
test_should_not_contain "<a href=\"/ipfs/$DIR_CID/\">..</a>" list_response
'
```

And now, this:
```go
{
 Name: "path gw: backlink on root CID should be hidden",
 Hint: `
 this test is written for the workshop, it will fail by default.
 But we can use it to show the rough idea of how to write tests.
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

There is an error in the test! You should see this in your error logs:

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

Try to find an error in the test, and fix it.

Once you're done, first, rebuild the docker image:

```bash
make docker
```

Then, run the tests again:

```bash
./gc ...
```

#### Implement a new test

> start from scratch and write a new test

Now, we're going to add a completely new test case to the same file (`tests/t0115_gateway_dir_listing_test.go`) and function (`TestGatewayDirListingOnPathGateway`).

The test should:
1. Make a request for `/ipfs/ROOT_CID/ą/ę`
1. Check if the status code is 301
1. Check if the location header is exactly `/ipfs/ROOT_CID/%c4%85/%c4%99` (WARNING: you're going to have to escape the `%` character as in `/ipfs/ROOT_CID/%%c4%%85/%%c4%%99`)

Here is how it looks like in sharness:
```
test_expect_success "path gw: redirect dir listing to URL with trailing slash" '
curl -sD - http://127.0.0.1:$GWAY_PORT/ipfs/${DIR_CID}/ą/ę > list_response &&
test_should_contain "HTTP/1.1 301 Moved Permanently" list_response &&
test_should_contain "Location: /ipfs/${DIR_CID}/%c4%85/%c4%99/" list_response
'
```

Once you're done, first, rebuild the docker image:

```bash
make docker
```

Then, run the tests again:

```bash
./gc ...
```

#### Implement a more complex test

> Introduce Headers, Body, etc.

Go to `tests/t0117_gateway_block_test.go`, look for TODOs and implement them.

### [optional] Fix a subdomain gateway test

To be able to run subdomain gateway tests, you need to configure your gateway to run on a subdomain.

You can use `kubo-config.example.sh`, for example.

You're going to need to restart the gateway after you change the config.

Then, when running tests, you can stop passing the `--specs -subdomain-gateway` flag.

You'll find a `subdomain-gateway` spec test that is failing (`TestGatewayDirListingOnSubdomainGateway`) in `tests/t0115_gateway_dir_listing_test.go`. It's marked with a TODO.

### [optional] Write a new spec test

Go to https://specs.ipfs.tech/http-gateways/, pick a spec, and write a test for it.

### TODOs

When we run the test suite it fails, what advice can we share fix this? I (laurent) run the tests in an IDE, it's easy to find there. If you run in CLI how do you fix it?
