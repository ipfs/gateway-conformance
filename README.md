<h1 align="center">
  <br>
  <a href="#readme"><img src="https://github.com/ipfs/gateway-conformance/assets/157609/4e7ba998-c7f7-415b-bd72-eef053474865" alt="Boxo logo" title="Boxo logo" width="300"></a>
  <br>
  Gateway Conformance
  <br>
</h1>

<p align="center" style="font-size: 1.2rem;">A set of GO and HTTP tools for testing implementation compliance with https://specs.ipfs.tech</p>

<p align="center">
  <a href="https://ipfs.tech"><img src="https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square" alt="Official Part of IPFS Project"></a>
  <a href="https://specs.ipfs.tech"><img src="https://img.shields.io/badge/specs-IPFS-blue.svg?style=flat-square" alt="IPFS Specifications"></a>
  <a href="https://github.com/ipfs/boxo/actions"><img src="https://img.shields.io/github/actions/workflow/status/ipfs/boxo/go-test.yml?branch=main" alt="ci"></a>
  <a href="https://github.com/ipfs/gateway-conformance/releases"><img alt="GitHub release" src="https://img.shields.io/github/v/release/ipfs/gateway-conformance?filter=!*rc*"></a>
</p>

<hr />

<!-- TOC -->

- [About](#about)
- [Usage](#usage)
  - [CLI](#cli)
  - [Docker](#docker)
  - [Github Action](#github-action)
  - [Web Dashboard](#web-dashboard)
- [Commands](#commands)
  - [Examples](#examples)
- [Releases](#releases)
- [Development](#development)
  - [Test DSL Syntax](#test-dsl-syntax)
- [License](#license)

<!-- /TOC -->

## About

Gateway Conformance test suite is a set of tools for testing implementation
compliance with a subset of [IPFS Specifications](https://specs.ipfs.tech). The
test suite is implementation and language-agnostic. Point `gateway conformance
test` at HTTP endpoint and specify which tests should run.

IPFS Shipyard uses it for ensuring specification compliance of the `boxo/gateway` library included in [Kubo](https://github.com/ipfs/kubo), [the most popular IPFS implementation](https://github.com/protocol/network-measurements/tree/master/reports),
that powers various [public gateways](https://ipfs.github.io/public-gateway-checker/), [IPFS Desktop](https://docs.ipfs.io/install/ipfs-desktop/), and [Brave](https://brave.com/ipfs-support/).


Some scenarios in which you may find this project helpful:

* You are building an product that relies on in-house IPFS Gateway and want to ensure HTTP interface is implemented correctly
* You are building an IPFS implementation and want to leverage existing HTTP test fixtures to tell if you are handling edge cases correctly
* You want to test if a trustless retrieval endpoint supports partial CARs from [IPIP-402](https://specs.ipfs.tech/ipips/ipip-0402/)
* You want to confirm a commercial service provider implemented content-addressing correctly

## Usage

The `gateway-conformance` can be run as a [standalone binary](#cli), a [Docker image](#docker), or a part of [Github Action](#github-actions).

Some of the tests require the tested gateway to be able to resolve specific fixtures CIDs or IPNS records.

Two high level [commands](/docs/commands.md) exist:
- [test](/docs/commands.md#test) (test runner with ability to specify a subset of tests to run)
- [extract-fixtures](/docs/commands.md#extract-fixtures) (allowing for custom provisioning of how test vectors are loaded into tested runtime)

### CLI

```console
$ # run subdomain-gateway tests against endpoint at http://localhost:8080 output as JSON
$ gateway-conformance test --gateway-url http://localhost:8080 --json report.json --specs +subdomain-gateway,-path-gateway -- -timeout 30m
```

If you are looking for TLDR, see [examples](/docs/examples.md).

### Docker

Prebuilt image at `ghcr.io/ipfs/gateway-conformance` can be used for both `test` and `extract-fixtures` commands:

```console
$ # extract fixtures to ./fixtures directory
$ docker run -v "${PWD}:/workspace" -w "/workspace" ghcr.io/ipfs/gateway-conformance:vA.B.C extract-fixtures --directory fixtures --merged false

$ # run subdomain-gateway tests against endpoint at http://localhost:8080
$ docker run --network host -v "${PWD}:/workspace" -w "/workspace" ghcr.io/ipfs/gateway-conformance:vA.B.C test --gateway-url http://localhost:8080 --json report.json --specs +subdomain-gateway,-path-gateway -- -timeout 30m
```

**NOTE:** replace `vA.B.C` with a [semantic version](https://github.com/ipfs/gateway-conformance/releases) version you want to test against

### Github Action

Common operations are possible via reusable GitHub actions:
- [`ipfs/gateway-conformance/.github/actions/test`](https://github.com/ipfs/gateway-conformance/blob/main/.github/actions/test/action.yml)
- [`ipfs/gateway-conformance/.github/actions/extract-fixtures`](https://github.com/ipfs/gateway-conformance/blob/main/.github/actions/extract-fixtures/action.yml)

To learn how to integrate them in the CI of your project, see real world examples in:
- [`kubo/../gateway-conformance.yml`](https://github.com/ipfs/kubo/blob/master/.github/workflows/gateway-conformance.yml) (fixtures imported into tested node)
- [`boxo/../gateway-conformance.yml`](https://github.com/ipfs/boxo/blob/main/.github/workflows/gateway-conformance.yml) (fixtures imported into a sidecar kubo node that is peered with tested library)
- [`bifrost-gateway/../gateway-conformance.yml`](https://github.com/ipfs/bifrost-gateway/blob/main/.github/workflows/gateway-conformance.yml) (fixtures imported into a kubo node that acts as a delegated block backend)

### Web Dashboard

Conformance test suite output can be plain text or JSON, which in turn can be
represented as a web dashboard which aggregates results from many test runs and
renders them on a static website.

The Implementation Dashboard instance at
[conformance.ipfs.tech](https://conformance.ipfs.tech/) is a view that
showcases some of well known and complete implementations of IPFS Gateways
in the ecosystem.

Learn more at [`/docs/web-dashboard.md`](/docs/web-dashboard.md)

## Commands

See `test` and `extract-fixtures` documentation at [`/docs/commands.md`](/docs/commands.md)

### Examples

Want to test mature specs, while disabling specific specs?
Or only test a specific spec (like trustless gateway), while disabling a sub-part of it (only blocks and CARS, no IPNS)?
See [`/docs/examples.md`](/docs/examples.md)

## Releases

The `main` branch may contain tests for features and IPIPs which are not yet
supported by stable releases of IPFS implementations.

Due to this, implementations SHOULD test themselves against a stable release
of this test suite instead.

See [`/releases`](https://github.com/ipfs/gateway-conformance/releases) for the list of available releases.

## Development

Want to improve the conformance test suite itself? 
See documentation at [`/docs/development.md`](/docs/development.md)

### Test DSL Syntax

Interested in write a new test case?
Test cases are written in Domain Specific Language (DLS) based on Golang. 
More details at [`/docs/test-dsl-syntax.md`](/docs/test-dsl-syntax.md)

## License

This project is dual-licensed under Apache 2.0 and MIT terms:

- Apache License, Version 2.0, ([LICENSE-APACHE](https://github.com/ipfs/kubo/blob/master/LICENSE-APACHE) or http://www.apache.org/licenses/LICENSE-2.0)
- MIT license ([LICENSE-MIT](https://github.com/ipfs/kubo/blob/master/LICENSE-MIT) or http://opensource.org/licenses/MIT)
