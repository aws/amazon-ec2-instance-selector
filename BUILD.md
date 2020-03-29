# Instance Selector: Build Instructions

## Install Go version 1.13+

There are several options for installing go:

1. If you're on mac, you can simply `brew install go`
2. If you'd like a flexible go installation manager consider using gvm https://github.com/moovweb/gvm
3. For all other situations use the official go getting started guide: https://golang.org/doc/install

## Compile

This project uses `make` to organize compilation, build, and test targets.

To compile cmd/main.go, which will build the full static binary and pull in depedent packages, run:
```
$ make compile
```

The resulting binary will be in the generated `build/` dir

```
$ make compile
/Users/$USER/git/amazon-ec2-instance-selector/
go build -a -ldflags "-X main.versionID=v0.9.0" -tags="aeislinux" -o /Users/$USER/git/amazon-ec2-instance-selector/build/ec2-instance-selector /Users/$USER/git/amazon-ec2-instance-selector/cmd/main.go

$ ls build/
ec2-instance-selector
```

## Test

You can execute the unit tests for the instance selector with `make`:

```
$ make unit-test
```

### Install Docker

The full test suite requires Docker to be installed. You can install docker from here: https://docs.docker.com/get-docker/

### Run All Tests

The full suite includes license-test, go-report-card, and more. See the full list in the [makefile](https://github.com/aws/amazon-ec2-instance-selector/blob/master/Makefile). NOTE: some tests require AWS Credentials to be configured on the system: 

```
$ make test
```

## Format

To keep our code readable with go conventions, we use `goimports` to format the source code.
Make sure to run `goimports` before you submit a PR or you'll be caught by our tests! 

You can use the `make fmt` target as a convenience
```
$ make fmt
```

## Generate all Platform Binaries

To generate binaries for all supported platforms (linx/amd64, linux/arm64, windows/amd64, etc.) run:

```
$ make build-binaries
```

The binaries are built using a docker container and are then `cp`'d out of the container and placed in `build/bin`

```
$ ls build/bin
ec2-instance-selector-darwin-amd64  ec2-instance-selector-linux-amd64   ec2-instance-selector-linux-arm     ec2-instance-selector-linux-arm64   ec2-instance-selector-windows-amd64
```

## See All Make Targets

To see all possible make targets and dependent targets, run:

```
$ make help
build-binaries: create-build-dir
build-docker-images:
build: create-build-dir compile
clean:
compile:
create-build-dir:
docker-build:
docker-push:
docker-run:
fmt:
go-report-card-test:
help:
image:
license-test:
output-validation-test: create-build-dir
push-docker-images:
readme-codeblock-test:
release: create-build-dir build-binaries build-docker-images push-docker-images upload-resources-to-github
spellcheck:
sync-readme-to-dockerhub:
test: spellcheck unit-test license-test go-report-card-test output-validation-test readme-codeblock-test
unit-test: create-build-dir
upload-resources-to-github:
version:
```
