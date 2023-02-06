# Gateway Conformance Testing Approach Proposal

## Key Concepts

- using JS UnixFS implementation to parse fixtures and get their CIDs, etc.
- decoupling fixtures provisioning from tests to make it easily replacable with custom provisioners
- defining test cases declratively in terms of request parameters and expected response
- using mocha-multi to support multiple reporters

## Issues

- [ ] writeable-gateway provisioner does not really work for uploading directories
- [ ] writeable-gateway provisioner will likely lack support for importer options

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
