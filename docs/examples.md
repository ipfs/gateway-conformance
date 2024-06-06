# Examples

- [Testing only mature specs](#testing-only-mature-specs)
- [Testing specific specs, regardless of their maturity level](#testing-specific-specs-regardless-of-their-maturity-level)
- [Testing mature specs and additionally enabling specific specs](#testing-mature-specs-and-additionally-enabling-specific-specs)
- [Testing mature specs, while disabling specific specs](#testing-mature-specs-while-disabling-specific-specs)
- [Testing specific spec (trustless gateway), while disabling a sub-part of it](#testing-specific-spec-trustless-gateway-while-disabling-a-sub-part-of-it)
- [Skip a specific test](#skip-a-specific-test)
- [Extracting the test fixtures](#extracting-the-test-fixtures)
- [Extracting the test fixtures into a single CAR file](#extracting-the-test-fixtures-into-a-single-car-file)

### Testing only mature specs

By default, all mature tests are run. Mature tests generally refer to specifications whose [status is mature](https://specs.ipfs.tech/meta/spec-for-specs/).

```bash
gateway-conformance test
```

### Testing specific specs, regardless of their maturity level

```bash
gateway-conformance test --specs subdomain-gateway,path-gateway
```

### Testing mature specs and additionally enabling specific specs

```bash
gateway-conformance test --specs +subdomain-gateway
```

### Testing mature specs, while disabling specific specs

```bash
gateway-conformance test --specs -subdomain-gateway,-dnslink-gateway
```

### Testing specific spec (trustless gateway), while disabling a sub-part of it

```bash
gateway-conformance test --specs trustless-gateway,-trustless-gateway-ipns
```

### Skip a specific test

Tests are skipped using Go's standard syntax:

```bash
gateway-conformance test -- -skip 'TestGatewayCar/GET_response_for_application/vnd.ipld.car/Header_Content-Length'
```

It supports regex:

```bash
gateway-conformance test -- -skip 'TestGatewayCar/*'
```

### Run only a specific test

Same syntax as for skipping:

```bash
gateway-conformance test -- -run 'TestGatewayCar/*'
```

### Extracting the test fixtures

```bash
gateway-conformance extract-fixtures
```

### Extracting the test fixtures into a single CAR file

```bash
gateway-conformance extract-fixtures --merged true
```

