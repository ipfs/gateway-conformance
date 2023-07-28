# gateway-conformance

`gateway-conformance` is a tool designed to test if an IPFS Gateway implementation complies with the IPFS Gateway Specification correctly. The tool is distributed as a Docker image, as well as a GitHub Action(s).

[![Conformance Production Dashboard](https://github.com/ipfs/gateway-conformance/actions/workflows/test-prod-e2e.yml/badge.svg?branch=master)]()

## Table of Contents

- [Commands](#commands)
  - [test](#test)
    - [Inputs](#inputs)
    - [Subdomain Testing and `subdomain-url`](#subdomain-testing-and-subdomain-url)
    - [Usage](#usage)
  - [extract-fixtures](#extract-fixtures)
    - [Inputs](#inputs-1)
    - [Outputs](#outputs)
    - [Usage](#usage-1)
- [Testing Your Gateway](#testing-your-gateway)
  - [Provisioning the Gateway](#provisioning-the-gateway)
- [Local Development](#local-development)
- [Examples](#examples)
- [APIs](#apis)
- [FAQ](#faq)
- [In Development](#in-development)

## Commands

### test

The `test` command is the main command of the tool. It is used to test a given IPFS Gateway implementation against the IPFS Gateway Specification.

#### Inputs

| Input | Availability | Description | Default |
|---|---|---|---|
| gateway-url | Both | The URL of the IPFS Gateway implementation to be tested. | http://localhost:8080 |
| subdomain-url | Both | The Subdomain URL of the IPFS Gateway implementation to be tested. | http://example.com |
| json | Both | The path where the JSON test report should be generated. | `./report.json` |
| xml | GitHub Action | The path where the JUnit XML test report should be generated. | `./report.xml` |
| html | GitHub Action | The path where the one-page HTML test report should be generated. | `./report.html` |
| markdown | GitHub Action | The path where the summary Markdown test report should be generated. | `./report.md` |
| specs | Both | A comma-separated list of specs to be tested. Accepts a spec (test only this spec), a +spec (test also this immature spec), or a -spec (do not test this mature spec). | Mature specs only |
| args | Both | [DANGER] The `args` input allows you to pass custom, free-text arguments directly to the Go test command that the tool employs to execute tests. | N/A |

##### Specs

By default, only mature specs (reliable, stable, or permanent) will be tested if this input is not provided. You can specify particular specs without any prefixes (e.g., subdomain-gateway, trustless-gateway, path-gateway) to test exclusively those, irrespective of their maturity status.

To selectively enable or disable specs based on their maturity, use the "+" and "-" prefixes. Adding a "+" prefix (e.g., +subdomain-gateway) means that the spec should be included in the test, in addition to the mature specs. Conversely, using a "-" prefix (e.g., -subdomain-gateway) means that the spec should be excluded from the test, even if it is mature.

If you provide a list containing both prefixed and unprefixed specs, the prefixed specs will be ignored. It is advisable to use either prefixed or unprefixed specs, but not both. However, you can include specs with both "+" and "-" prefixes in the same list.

##### Args

This input should be used sparingly and with caution, as it involves interacting with the underlying internal processes, which may be subject to changes. It is recommended to use the `args` input only when you have a deep understanding of the tool's inner workings and need to fine-tune the testing process. Users should be mindful of the potential risks associated with using this input.

#### Subdomain Testing and `subdomain-url`

The `subdomain-url` parameter is utilized when testing subdomain support in your IPFS gateway. It can be set to any domain that your gateway permits.
During testing, the suite keeps connecting to the `gateway-url` while employing HTTP techniques to simulate requests as if they were sent to the subdomain.
This approach enables testing of local gateways during development or continuous integration (CI) scenarios.

A few examples:

| Use Case | gateway-url | subdomain-url |
|----------|-------------|---------------|
| CI & Dev   | http://127.0.0.1:8080 | http://example.com |
| Production | https://dweb.link     | https://dweb.link  |

#### Usage

##### GitHub Action

```yaml
- name: Run gateway-conformance tests
  uses: ipfs/gateway-conformance/.github/actions/test@v1
  with:
    gateway-url: http://localhost:8080
    specs: +subdomain-gateway,-path-gateway
    json: report.json
    xml: report.xml
    markdown: report.md
    html: report.html
    args: -timeout 30m
```

##### Docker

```bash
docker run --network host -v "${PWD}:/workspace" -w "/workspace" ghcr.io/ipfs/gateway-conformance test --gateway-url http://localhost:8080 --json report.json --specs +subdomain-gateway,-path-gateway -- -timeout 30m
```

### extract-fixtures

The `extract-fixtures` command is used to extract the test fixtures from the `gateway-conformance` tool.

#### Inputs

| Input | Availability | Description | Default |
|---|---|---|---|
| output | Both | The path where the test fixtures should be extracted. | `./fixtures` |
| merged | Both | Whether the fixtures should be merged into as few files as possible. | `false` |

#### Outputs

When you use `--merged=true` the following files are be generated:

- `fixtures.car`: A car file that contains all the blocks required to run the tests
- `dnslinks.json`: A configuration file listing all the dnslink names required to run the tests related to DNSLinks
- `*.ipns-record`: Many raw ipns-record files required to run the tests related to IPNS

Examples of how to import these in Kubo are shown in [`kubo-config.example.sh`](./kubo-config.example.sh) and the [`Makefile`](./Makefile).

Without `--merged=true`, many car files and dnslink configurations file will be generated, we don't recommend using these.

#### Usage

##### GitHub Action

```yaml
- name: Extract gateway-conformance fixtures
  uses: ipfs/gateway-conformance/.github/actions/extract-fixtures@v1
  with:
    output: fixtures
    merged: false
```

##### Docker

```bash
docker run -v "${PWD}:/workspace" -w "/workspace" ghcr.io/ipfs/gateway-conformance extract-fixtures --output fixtures --merged false
```

## Testing Your Gateway

You can find the workflow that runs the gateway conformance test suite against Kubo in the file [.github/workflows/test.yml](.github/workflows/test.yml). This can serve as a good starting point when setting up your own test suite.

We've also aimed to keep the [kubo-config.example.sh](kubo-config.example.sh) script and the [Makefile](Makefile) as straightforward as possible to provide useful examples to get started.

### Provisioning the Gateway

We make minimal assumptions about the capabilities of the gateway being tested. Which means that we don't require nor expect the gateway to be writable. Therefore, you need to provision the gateway with test fixtures before running the test suite.

These fixtures are located in the `./fixtures` folders. We distribute tools for extracting them. Refer to the documentation for the `extract-fixtures` command for more details.

**Fixtures:**

- Blocks & Dags: These are served as [CAR](https://ipld.io/specs/transport/car/) file(s).
- IPNS Records: These are distributed as files containing [IPNS Record](https://specs.ipfs.tech/ipns/ipns-record/#ipns-record) [serialized as protobuf](https://specs.ipfs.tech/ipns/ipns-record/#record-serialization-format). The file name includes the Multihash of the public key ([IPNS Name](https://specs.ipfs.tech/ipns/ipns-record/#ipns-name)) in this format: `pubkey(_optional_suffix)?.ipns-record`. We may decide to [share CAR files](https://github.com/ipfs/specs/issues/369) in the future.
- DNS Links: These are distributed as `yml` configurations. You can use the `--merge` option to generate a consolidated `.json` file, which can be more convenient for use in a shell script.

## Local Development

This is how we use the test-suite when we work on the suite itself or a gateway implementation:

```sh
# Generate the fixtures
make fixtures.car

# Import the fixtures in Kubo
# We import car files and ipns-records during this step.
# We import DNSLink fixtures through the `IPFS_NS_MAP` below. 
make provision-kubo

# Configure Kubo for the test-suite
# This also generated the `IPFS_NS_MAP` variable to setup DNSLink fixtures
source ./kubo-config.example.sh

# Start a Kubo daemon in offline mode
ipfs daemon --offline
```

By then the gateway is configured and you may run the test-suite.

```sh
make test-kubo

# run with subdomain testing which requires more configuration (see kubo-config.example.sh)
make test-kubo-subdomains
```

If you are using a different gateway and would like to use a different configuration, the [Makefile](./Makefile) and configuration scripts are great, up-to-date, starting points.

The test-suite is a regular go test-suite, which means that any IDE integration will work as-well.
You can use env variables to configure the tests from your IDE.

Here is an example for VSCode, `example.com` is the domain configured via [kubo-config.example.sh](./kubo-config.example.sh)

```json
{
  "go.testEnvVars": {
    "GATEWAY_URL": "http://127.0.0.1:8080",
    "SUBDOMAIN_GATEWAY_URL": "http://example.com",
    "GOLOG_LOG_LEVEL": "conformance=debug"
  },
}
```

With this configuration, the tests will appear in `Testing` on VSCode's left sidebar.

It's also possible to run test suite locally and use `make ./reports/output.html` to generate a human-readable report from the test results in `reports/output.json`.

## Examples

The examples are going to use `gateway-conformance` as a wrapper over `docker run -v "${PWD}:/workspace" -w "/workspace" ghcr.io/ipfs/gateway-conformance` for simplicity.

### Testing only mature specs

By default, all mature tests are run. Mature tests generally refer to specifications whose [status is mature](https://specs.ipfs.tech/meta/spec-for-specs/).

```bash
gateway-conformance test
```

### Testing specific specs, regardless of their maturity level

```bash
gateway-conformance test --specs subdomain-gateway,path-gateway
```

### Testing mature specs and additionally enabling specific specs

```bash
gateway-conformance test --specs +subdomain-gateway
```

### Testing mature specs, while disabling specific specs

```bash
gateway-conformance test --specs -subdomain-gateway,-dnslink-gateway
```

### Testing specific spec (trustless gateway), while disabling a sub-part of it

```bash
gateway-conformance test --specs trustless-gateway,-trustless-gateway-ipns
```

### Skip a specific test

Tests are skipped using Go's standard syntax:

```bash
gateway-conformance test -skip 'TestGatewayCar/GET_response_for_application/vnd.ipld.car/Header_Content-Length'
```

### Extracting the test fixtures

```bash
gateway-conformance extract-fixtures
```

### Extracting the test fixtures into a single CAR file

```bash
gateway-conformance extract-fixtures --merged true
```

## APIs

### Templating

golang's default string formating package is similar to C. Format strings might look like `"this is a %s"` where `%s` is a verb that will be replaced at runtime.

These verbs collides with URL-escaping a lot, strings like `/ipfs/Qm.../%c4%85/%c4%99` might trigger weird errors. We implemented a minimal templating library that is used almost everywhere in the test.

It uses `{{name}}` as a replacement for `%s`. Other verbs are not supported.


```golang
Fmt("{{action}} the {{target}}", "pet", "cat") // => "pet the cat"
```

Backticks enable use of verbatim strings, without having to deal with golang-specific escaping of things like double quotes:

```golang
Fmt(`Etag: W/"{{etag-value}}"`, "weak-key") // => "ETag: W/\"weak-key\""
```

It is required to always provide a meaningful `{{name}}`:

```golang
Fmt(`/ipfs/{{cid}}/%c4%85/%c4%99`, fixture.myCID) // => "/ipfs/Qm..../%c4%85/%c4%99"
```

Values are replaced in the order they are defined, and you may reuse named values

```golang
Fmt(`<a href="{{cid}}">{{label}}}</a><a href="{{cid}}/index.html">index</a>`, fixture.myCID, "Link Title!") // => '<a href="Qm...">Link Title!</a><a href="Qm..../index.html">index</a>'
```

You may escape `{{}}` by using more than two opening or closing braces,

```golang
Fmt("{foo}") // => "{foo}"
Fmt("{{{foo}}}") // => "{{foo}}"
Fmt("{{{{foo}}}}") // => "{{{foo}}}"
Fmt("{{{foo}}}") // => {{foo}}
```

This templating is used almost everywhere in the test sugar, for example in request Path:

```golang
Request().Path("ipfs/{{cid}}", myCid) // will use "ipfs/Qm...."
```

## FAQ

### How to generate XML, HTML and Markdown reports when using the tool as a Docker container?

The tool can generate XML, HTML and Markdown reports when used as a GitHub Action. However, when using the tool as a Docker container, you can generate these reports by using the [`saxon` Docker image](https://github.com/pl-strflt/saxon). You can draw inspiration from the [gotest-json-to-junit-xml](https://github.com/pl-strflt/gotest-json-to-junit-xml) and the [junit-xml-to-html](https://github.com/pl-strflt/junit-xml-to-html) GitHub Actions.

Please let us know if you would like to see this feature implemented directly in the Docker image distribution.

## In Development

- How to deal with subdomains & configuration (t0114 for example)?
  - Some test relies on querying URLs like `http://$CIDv1.ipfs.example.com/`. While `http://$CIDv1.ipfs.localhost/` works by default, do we need / want to test with `.example.com`?
- Debug logging
  - Set the environment variable `GOLOG_LOG_LEVEL="conformance=debug"` to toggle debug logging.