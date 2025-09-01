# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.8.3] - 2025-09-01
### Changed
- Added temporary support for legacy 200 with X-Stream-Error header behavior when missing blocks are detected during CAR streaming. This ensures compatibility with current boxo/gateway implementations that defer header setting. This third option will be removed once implementations ship with [ipfs/boxo#1019](https://github.com/ipfs/boxo/pull/1019). [#245](https://github.com/ipfs/gateway-conformance/pull/245)

## [0.8.2] - 2025-08-31
### Changed
- Relaxed CAR tests to accept both HTTP 200 and 404 for non-existing paths. The response code now depends on implementation details such as locality and cost of path traversal checks. Implementations that can efficiently detect non-existing paths should return 404 (improved behavior per [ipfs/boxo#458](https://github.com/ipfs/boxo/issues/458)). Implementations focusing on stateless streaming and low latency may return 200 with partial CAR up to the missing link (legacy behavior). [#244](https://github.com/ipfs/gateway-conformance/pull/244)

## [0.8.1] - 2025-06-17
### Changed
- DAG-CBOR HTML preview pages previously had to be returned without Cache-Control headers. Now they can use Cache-Control headers similar to those used in generated UnixFS directory listings. [#241](https://github.com/ipfs/gateway-conformance/pull/241)

## [0.8.0] - 2025-05-28
### Changed
- Comprehensive tests for HTTP Range Requests over deserialized UnixFS files have been added. The `--specs path-gateway` now requires support for at least single-range requests. Deserialized range-requests can be skipped with `--skip 'TestGatewayUnixFSFileRanges'` [#213](https://github.com/ipfs/gateway-conformance/pull/213)
- Updated dependencies [#236](https://github.com/ipfs/gateway-conformance/pull/236) & [#239](https://github.com/ipfs/gateway-conformance/pull/239)

## [0.7.1] - 2025-01-03
### Changed
- Expect all URL escapes to use uppercase hex [#232](https://github.com/ipfs/gateway-conformance/pull/232)

## [0.7.0] - 2025-01-03
### Changed
- Update dependencies [#226](https://github.com/ipfs/gateway-conformance/pull/226) and [#227](https://github.com/ipfs/gateway-conformance/pull/227)
- Expect upper-case hex digits in escaped redirect URL [#225](https://github.com/ipfs/gateway-conformance/pull/225)

## [0.6.2] - 2024-08-09
### Changed
- Relaxed negative test of TAR response [#221](https://github.com/ipfs/gateway-conformance/pull/221)

## [0.6.1] - 2024-07-29
### Changed
- Support meaningful `Cache-Control` on generated UnixFS directory listing responses on `/ipfs` namespace

## [0.6.0] - 2024-06-10
### Changed
- Gateway URL
  - `--gateway-url` is no longer defaulting to predefined URL. User has to
    provide it via CLI or `GATEWAY_URL` environment variable or the test suite
    will refuse to start.
  - This aims to ensure no confusion about which gateway endpoint is being
    tested.
  - Docs and examples use `--gateway-url http://127.0.0.1:8080` to ensure no
    confusion with `localhost:8080` subdomain gateway feature in IPFS
    implementations like Kubo.
- Subdomain URL and UX related to subdomain tests
  - The `--subdomain-url` is no longer set by default.
  - User has to provide the origin of the subdomain gateway via CLI or
    `SUBDOMAIN_GATEWAY_URL` to be used during subdomain tests. This aims to
    ensure no confusion about which domain name is being tested.
  - Simplified the way `--subdomain-url` works. We no longer run implicit tests
    against `http://localhost` in addition to the URL passed via
    `--subdomain-url`. To test more than one domain, run test multiple times.
  - `localhost` subdomain gateway tests  are no longer implicit. To run tests
    against `localhost` use `--subdomain-url http://localhost:8080`
-  DNSLink test fixtures changed
   - DNSLink fixtures no longer depend on `--subdomain-url` and use unrelated
     `*.example.org` domains instead.
   - `gateway-conformance extract-fixtures` creates `dnslinks.IPFS_NS_MAP` with
     content that can be directly set as `IPNS_NS_MAP` environment variable
     supported by various implementations, incl.
     [Kubo](https://github.com/ipfs/kubo/blob/master/docs/environment-variables.md#ipfs_ns_map)
     and
     [Rainbow](https://github.com/ipfs/rainbow/blob/main/docs/environment-variables.md#ipfs_ns_map).
- Docker: image can now be run under non-root user
- HTTP Proxy tests are no longer implicit. An explicit spec named
  `proxy-gateway` exists now, and can be disabled via `--specs -proxy-gateway`.

## [0.5.2] - 2024-05-20
### Changed
- Fixed: relaxed dag-cbor error check ([#205](https://github.com/ipfs/gateway-conformance/pull/205))
- Fixed: Header().Has works properly for checking multiple values ([#207](https://github.com/ipfs/gateway-conformance/pull/207))

## [0.5.1] - 2024-04-11
- Removed byte range text for DAG-CBOR objects converted to `text/html`. [PR](https://github.com/ipfs/gateway-conformance/pull/202)

## [0.5.0] - 2024-01-25
### Changed
- Fixed tests of CAR requests with `entity-bytes` and negative indexing. [PR](https://github.com/ipfs/gateway-conformance/pull/190) (BREAKING CHANGE)
- Fixed IPNS provisioning with Kubo. [PR](https://github.com/ipfs/gateway-conformance/pull/192)

## [0.4.2] - 2023-11-20
### Changed
- Fixed versioning in Docker containers. [PR](https://github.com/ipfs/gateway-conformance/pull/179)

## [0.4.1] - 2023-10-11
### Changed
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
