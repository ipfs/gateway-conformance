# Gateway Conformance Testing Approach Proposal

## Key Concepts

- TBD

## Issues

- TBD

## Usage

### Retrieve fixtures

```bash
docker run -v "${PWD}:/workspace" ghcr.io/ipfs/gateway-conformance extract-fixtures /workspace/[output-dir]
```

### Run tests

```bash
docker run --network host -v "${PWD}:/workspace" ghcr.io/ipfs/gateway-conformance test [gateway-url] /workspace/[output-xml]
```

### Generate an html report

```bash
docker run --rm -v "${PWD}:/workspace" ghcr.io/pl-strflt/junit-xml-to-html:latest no-frames /workspace/[output-xml] /workspace/[output-html]
```

### Generate a single car file for testing

```bash
docker run --network host -w "/workspace" -v "${PWD}:/workspace" ghcr.io/ipfs/gateway-conformance merge-fixtures /workspace/[output-car]
```
