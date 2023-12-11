# `gateway-conformance` Commands

- [Commands](#commands)
  - [test](#test)
    - [Inputs](#inputs)
      - [Specs](#specs)
      - [Args](#args)
    - [Subdomain Testing and `subdomain-url`](#subdomain-testing-and-subdomain-url)
    - [Usage](#usage)
      - [GitHub Action](#github-action)
      - [Docker](#docker)
  - [extract-fixtures](#extract-fixtures)
    - [Inputs](#inputs-1)
    - [Outputs](#outputs)
    - [Usage](#usage-1)
      - [GitHub Action](#github-action-1)
      - [Docker](#docker-1)
- [Testing Your Gateway](#testing-your-gateway)
  - [Provisioning the Gateway](#provisioning-the-gateway)
- [Local Development](#local-development)
- [Examples](#examples)

## Commands

### test

The `test` command is the main command of the tool. It is used to test a given IPFS Gateway implementation against the [IPFS Gateway Specification](https://specs.ipfs.tech/http-gateways/).

#### Inputs

| Input | Availability | Description | Default |
|---|---|---|---|
| gateway-url | Both | The URL of the IPFS Gateway implementation to be tested. | http://localhost:8080 |
| subdomain-url | Both | The URL to be used in Subdomain feature tests based on Host HTTP header. | http://example.com |
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

The `subdomain-url` parameter is utilized when testing subdomain support in your IPFS gateway. It can be set to any domain that your gateway safelisted for Subdomain feature.
During testing, the suite sends HTTP requests to the `gateway-url` while setting HTTP `Host` header to simulate Subdomain requests.
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

## Examples

See [`examples.md`](./examples.md)

