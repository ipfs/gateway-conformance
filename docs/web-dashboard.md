# Web Dashboard

- [Summary](#summary)
- [How it works](#how-it-works)
  - [Building the dashboard](#building-the-dashboard)
  - [Local build of the dashboard](#local-build-of-the-dashboard)
  - [Adding new implementation to the dashboard](#adding-new-implementation-to-the-dashboard)

## Summary

Gateway Conformance test suite output can be represented as a web dashboard
which aggregates results from many test runs and renders them on a static
website.

IPFS Project uses the web dashboard instance at
[conformance.ipfs.tech](https://conformance.ipfs.tech/) for tracking selected
reference implementations, but everyone is free to fork this repository and use
own instance for internal purposes, if needed.

## How it works

For every implementations that have been added to the
[`REPOSITORIES`](../REPOSITORIES) file, our dashboard generation workflow loads
the most recent `gateway-conformance` test results from their CI. You can try
this locally with the `make website` command.

### Building the dashboard

- Set up a GitHub Token: Ensure you have a GitHub Token with `repo:read` scope.
  This is required to download artifacts from other repositories.
- Run the Build Command: `GH_TOKEN=your-gh-token make website`

This command downloads the latest test artifacts from the repositories listed
in the `REPOSITORIES` file. Then it generates a static website with Hugo in the
`www/public` directory.

### Local build of the dashboard

- Use `make website` to generate all the assets required to build the static dashboard
- Use `cd ./www && hugo server` to start a local server with live-reload
- Use `cd ./www/themes/conformance && npm run build` to re-build the theme's styles

### Adding new implementation to the dashboard

The dashboard is hosted at
[conformance.ipfs.tech](https://conformance.ipfs.tech/). It aggregates test
outputs from various IPFS implementations and renders them on a static website.

To add a new implementation to the dashboard:

- Ensure that implementation has significant user base and passes conformance tests.
- Ensure that repository that runs conformance tests is public.
- Ensure the CI in the repository generates and uploads the `output.json` artifact file.
  - For example, see [`gateway-conformance.yml` file Kubo](https://github.com/ipfs/kubo/blob/master/.github/workflows/gateway-conformance.yml).
- Open PR in this repository to add repository name to the [REPOSITORIES](../REPOSITORIES) file.
- Once merged, test results from the new artifact will be picked up
  automatically and the new implementation will show up on the dashboard.
