# Gateway Conformance Testing Approach Proposal

## Key Concepts

- TBD

## Issues

- TBD

## Usage

### Retrieve fixtures

```bash
docker run -w "/workspace" -v "${PWD}:/workspace" ghcr.io/ipfs/gateway-conformance extract-fixtures [output-dir]
```

### Run tests

```bash
docker run --network host -w "/workspace" -v "${PWD}:/workspace" ghcr.io/ipfs/gateway-conformance test [gateway-url] [output-xml]
```

### Generate an html report

```bash
docker run --rm -w "/workspace" -v "${PWD}:/workspace" ghcr.io/pl-strflt/junit-xml-to-html:latest no-frames [output-xml] [output-html]
```

### Generate a single car file for testing

```bash
docker run --network host -w "/workspace" -v "${PWD}:/workspace" ghcr.io/ipfs/gateway-conformance merge-fixtures # outputs to ./fixtures.car
```
