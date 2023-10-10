# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
- Loosened the `Cache-Control` and `Last-Modified` checks for IPNS paths, as they are now allowed. [PR](https://github.com/ipfs/gateway-conformance/pull/173)

## [0.4.0] - 2023-10-02
### Added
- Added tests for HTTP Range requests, as well as some basic helpers for `AnyOf` and `AllOf`. [PR](https://github.com/ipfs/gateway-conformance/pull/162)

## [0.3.1] - 2023-09-15
### Added
- Specs Dashboard Output. [PR](https://github.com/ipfs/gateway-conformance/pull/163)
- `--version` flag shows the current version
- Metadata logging used to associate tests with custom data like versions, specs identifiers, etc.
- Output Github's workflow URL with metadata. [PR](https://github.com/ipfs/gateway-conformance/pull/145)
- Basic Dashboard Output with content generation. [PR](https://github.com/ipfs/gateway-conformance/pull/152)
- Test Group Metadata on Tests. [PR](https://github.com/ipfs/gateway-conformance/pull/156)
- Specs Metadata on Tests. [PR](https://github.com/ipfs/gateway-conformance/pull/159)

### Changed
- Escape test names to avoid confusion when processing test hierarchies. [PR](https://github.com/ipfs/gateway-conformance/pull/166)

## [0.3.0] - 2023-07-31
### Added
- `--verbose` flag displays all the output to the console
- `Expect.Headers.ChecksAll`: an expectation to test all the header values (0, 1, or more)

### Changed
- finalized port of Kubo's sharness tests. [PR](https://github.com/ipfs/gateway-conformance/pull/92)
- `extract-fixtures --merged` generates a car version 1 with a single root now
- refactored multi-range requests. [PR](https://github.com/ipfs/gateway-conformance/pull/113)

## [0.2.0] - 2023-06-26
### Added
- `carFixture.MustGetChildren`
- Gateway backend timeout test for entity-bytes from IPIP-402. [Issue](https://github.com/ipfs/gateway-conformance/issues/75).

### Changed
- Renamed methods using `Children` into `Descendants` when relevant
- CAR tests no longer check for the roots. See discussion in [IPIP-402](https://github.com/ipfs/specs/pull/402).

## [0.1.0] - 2023-06-08
### Added
- `Fmt` a string interpolation that replaces golang's and works better with HTML entities, and HTTP headers and URLs.
- Support for calling multiple requests in a single test case and comparing their payloads.

### Changed
- `Path(url)` does not add a leading `/` to the URL anymore.
- Do not follow redirects by default anymore, remove `DoNotFollowRedirect` and add `FollowRedirect`.
- `Body` check is running in its own test. #67

## [0.0.2] - 2023-06-01
### Removed
- Body check for subdomain redirection

## [0.0.1] - 2023-03-27
### Added
- v0 of the Gateway Conformance test suite
