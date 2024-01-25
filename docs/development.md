# Development

- [The need for provisioning fixtures](#the-need-for-provisioning-fixtures)
- [Developing against Kubo](#developing-against-kubo)
  - [Provisioning local instance](#provisioning-local-instance)
- [FAQ](#faq)
  - [How to enable debug logging](#how-to-enable-debug-logging)

## The need for provisioning fixtures

We make minimal assumptions about the capabilities of the gateway being tested.
Which means that we don't require nor expect the gateway to be writable.
Therefore, you need to provision the gateway with test fixtures before running
the test suite.

These fixtures are located in the `./fixtures` folders. We distribute tools for
extracting them. Refer to the documentation for the `extract-fixtures` command
for more details.

**Fixtures:**

- Blocks & Dags: These are served as [CAR](https://ipld.io/specs/transport/car/) file(s).
- IPNS Records: These are distributed as files containing [IPNS Record](https://specs.ipfs.tech/ipns/ipns-record/#ipns-record) [serialized as protobuf](https://specs.ipfs.tech/ipns/ipns-record/#record-serialization-format). The file name includes the Multihash of the public key ([IPNS Name](https://specs.ipfs.tech/ipns/ipns-record/#ipns-name)) in this format: `pubkey(_optional_suffix)?.ipns-record`. We may decide to [share CAR files](https://github.com/ipfs/specs/issues/369) in the future.
- DNSLinks: These are distributed as `yml` configurations. You can use the `--merge` option to generate a consolidated `.json` file, which can be more convenient for use in a shell script.

## Developing against Kubo

When working on new tests, the easiest way to provision is to run against
boxo/gateway implementation that ships with Kubo.

You can find the workflow that runs the gateway conformance test suite against
Kubo in the file
[.github/workflows/test-kubo-e2e.yml](.github/workflows/test-kubo-e2e.yml).
This can serve as a good starting point when setting up your own test suite.

We also aim to keep the [kubo-config.example.sh](kubo-config.example.sh)
script and the [Makefile](Makefile) as straightforward as possible to provide
useful examples to get started.

### Provisioning local instance

This is how we use the test-suite when we work on the suite itself or a gateway implementation:

```sh
# Generate the fixtures
make fixtures.car

# Import the fixtures in Kubo
# We import car files and ipns-records during this step.
# We import DNSLink fixtures through the `IPFS_NS_MAP` below. 
make provision-kubo

# Configure Kubo for the test-suite
# This also generated the `IPFS_NS_MAP` variable to setup DNSLink fixtures
source ./kubo-config.example.sh

# Start a Kubo daemon in offline mode
ipfs daemon --offline
```

By then the gateway is configured and you may run the test-suite.

```sh
make test-kubo

# run with subdomain testing which requires more configuration (see kubo-config.example.sh)
make test-kubo-subdomains
```

If you are using a different gateway and would like to use a different configuration, the [Makefile](./Makefile) and configuration scripts are great, up-to-date, starting points.

The test-suite is a regular go test-suite, which means that any IDE integration will work as-well.
You can use env variables to configure the tests from your IDE.

Here is an example for VSCode, `example.com` is the domain configured via [kubo-config.example.sh](./kubo-config.example.sh)

```json
{
  "go.testEnvVars": {
    "GATEWAY_URL": "http://127.0.0.1:8080",
    "SUBDOMAIN_GATEWAY_URL": "http://example.com",
    "GOLOG_LOG_LEVEL": "conformance=debug"
  },
}
```

With this configuration, the tests will appear in `Testing` on VSCode's left sidebar.

It's also possible to run test suite locally and use `make ./reports/output.html` to generate a human-readable report from the test results in `reports/output.json`.

## FAQ

### How to enable debug logging

Set the environment variable `GOLOG_LOG_LEVEL="conformance=debug"` to toggle debug logging.

### How to make a new release

Create a new PR that modifies CHANGELOG.md,
see [changelog-driven-release#how-it-works](https://github.com/ipdxco/changelog-driven-release#how-it-works) for more details.
