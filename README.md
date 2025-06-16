****# Platform Mesh - golang-commons
![Build Status](https://github.com/platform-mesh/golang-commons/actions/workflows/pipeline.yml/badge.svg)
[![REUSE status](
https://api.reuse.software/badge/github.com/platform-mesh/golang-commons)](https://api.reuse.software/info/github.com/platform-mesh/golang-commons)
## Description

golang-commons contains golang library packages to be reused across microservices and operators/controllers. The scope includes, but is not limited to:

- JWT handling
- Context handling
- Error handling
- Logging
- [Controllers](./controller/README.md)

## Getting started

Add the dependency to your go module based project like so:

```
go get github.com/platform-mesh/golang-commons
```

## Releasing

The release is performed automatically through a GitHub Actions Workflow.

All the released versions will be available through access to GitHub (as any other Golang library).

## Requirements

golang-commons requires a installation of go. Checkout the [go.mod](go.mod) for the required go version.

## Mocks

golang commons uses mockery for mocking. If mock is absent, you can update `.mockery.yaml` file by adding the missing mock. Then run the following command to generate the mock files:
```
task mockery
``` 
P.S. If you have golang installed, it automatically installs the mockery binary in `golang-commons/bin` directory.

## Contributing

Please refer to the [CONTRIBUTING.md](CONTRIBUTING.md) file in this repository for instructions on how to contribute to Platform Mesh.

## Code of Conduct

Please refer to the [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) file in this repository informations on the expected Code of Conduct for contributing to Platform Mesh.

## Licensing

Copyright 2024 SAP SE or an SAP affiliate company and Platform Mesh contributors. Please see our [LICENSE](LICENSE) for copyright and license information. Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/platform-mesh/golang-commons).
