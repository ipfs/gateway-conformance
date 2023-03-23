# gateway-conformance

`gateway-conformance` is a tool designed to test if an IPFS Gateway implementation complies with the IPFS Gateway Specification correctly. The tool is distributed as a Docker image, as well as a GitHub Action(s).

## Table of Contents

- [Commands](#commands)
  - [test](#test)
    - [Inputs](#inputs)
    - [Usage](#usage)
  - [extract-fixtures](##extract-fixtures)
    - [Inputs](#inputs-1)
    - [Usage](#usage-1)
- [Examples](#examples)
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
| merged | Both | Whether the fixtures should be merged into a single CAR file. | `false` |

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

## Examples

The examples are going to use `gateway-conformance` as a wrapper over `docker run -v "${PWD}:/workspace" -w "/workspace" ghcr.io/ipfs/gateway-conformance` for simplicity.

### Testing only mature specs

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
gateway-conformance test --specs -subdomain-gateway
```

### Extracting the test fixtures

```bash
gateway-conformance extract-fixtures
```

### Extracting the test fixtures into a single CAR file

```bash
gateway-conformance extract-fixtures --merged true
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