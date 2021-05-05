#!/bin/bash
set -euo pipefail

SCRIPTPATH="$( cd "$(dirname "$0")" ; pwd -P )"
BUILD_BIN="$SCRIPTPATH/../../build/bin"

OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=amd64
BINARY_NAME="ec2-instance-selector-$OS-$ARCH"
LICENSE_TEST_TAG="aeis-license-test"

SUPPORTED_PLATFORMS="$OS/$ARCH" make -f $SCRIPTPATH/../../Makefile build-binaries

SUPPORTED_PLATFORMS="$OS/$ARCH" make -s -f $SCRIPTPATH/../../Makefile build-binaries
docker build --build-arg=GOPROXY=direct -t $LICENSE_TEST_TAG $SCRIPTPATH/
docker run -i -e GITHUB_TOKEN --rm -v $SCRIPTPATH/:/test -v $BUILD_BIN/:/aeis-bin $LICENSE_TEST_TAG golicense /test/license-config.hcl /aeis-bin/$BINARY_NAME
