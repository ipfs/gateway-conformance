# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Changed
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
