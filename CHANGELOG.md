# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased
### Added
- `--version` flag shows the current version
- Metadata logging used to associate tests with custom data like versions, specs identifiers, etc.
- Test Group Metadata on Tests
- IPIP Metadata on Tests and SugarTest

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
