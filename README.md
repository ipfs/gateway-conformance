# Gateway Conformance Testing Approach Proposal

## Key Concepts

- Built with Go for stability & ease of building
- TBD

## Issues / Open Questions

- How to deal with subdomains & configuration (t0114 for example)?
  - Some test relies on querying URLs like `http://$CIDv1.ipfs.example.com/`. While `http://$CIDv1.ipfs.localhost/` works by default, do we need / want to test with `.example.com`?

## Usage

### Retrieve fixtures

```bash
docker run -v "${PWD}:/workspace" -w "/workspace" ghcr.io/ipfs/gateway-conformance extract-fixtures [OUTPUT_DIR]
```

### Run tests

```bash
docker run --network host -v "${PWD}:/workspace" -w "/workspace" ghcr.io/ipfs/gateway-conformance test [gateway-url] [OUTPUT_JSON]
```

### Generate a XML report

```bash
docker run --rm -v "${PWD}:/workspace" -w "/workspace" --entrypoint "/bin/bash" ghcr.io/pl-strflt/saxon:v1 -c """
  java -jar /opt/SaxonHE11-5J/saxon-he-11.5.jar -s:<(jq -s '.' [OUTPUT_JSON]) -xsl:/etc/gotest.xsl -o:[OUTPUT_XML]
"""

```

### Generate a HTML report

```bash
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/pl-strflt/saxon:v1 -s:[OUTPUT_XML] -xsl:/etc/junit-noframes-saxon.xsl -o:[OUTPUT_HTML]
```

### Generate a single car file for testing

```bash
docker run --network host -w "/workspace" -v "${PWD}:/workspace" ghcr.io/ipfs/gateway-conformance merge-fixtures /workspace/[output-car]
```
