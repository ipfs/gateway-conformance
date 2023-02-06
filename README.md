# Gateway Conformance Testing Approach Proposal

## Key Concepts

- using JS UnixFS implementation to parse fixtures and get their CIDs, etc.
- decoupling fixtures provisioning from tests to make it easily replacable with custom provisioners
- defining test cases declratively in terms of request parameters and expected response
- using mocha-multi to support multiple reporters

## Issues

- [ ] writeable-gateway provisioner does not really work for uploading directories
- [ ] writeable-gateway provisioner will likely lack support for importer options
- [ ] doc and markdown reporters try to include test code directly in the reports which does not really work with declarative tests because the code blocks end up saying `[native code]` only
- [ ] markdown reporter does not report on errors at all

## Usage

### Install dependencies

```bash
npm ci
```

### Patch dependencies

```bash
npm run patch
```

### Provision fixtures

```bash
npm run provision <kubo|writeable-gateway> [<dir>]
```

### Run tests

```bash
npm test
```

## Dependencies

- uses https://marc-ed-raffalli.github.io/declarative-e2e-test/ to define test suites declaratively
- uses https://github.com/ipfs/js-ipfs-unixfs to parse fixtures and get their CIDs, etc.
- uses https://mochajs.org/ as a test framework (easily replacable with others)
- uses https://www.npmjs.com/package/mocha-multi to support multiple reporters
- uses built-in mocha reporters (spec, doc, markdown, xml, json) to report on test results (easily replacable with custom reporters)
