VERSION ?= $(shell git describe --tags --always --dirty)
BIN ?= ec2-instance-selector
IMG ?= amazon/amazon-ec2-instance-selector
REPO_FULL_NAME ?= aws/amazon-ec2-instance-selector
IMG_TAG ?= ${VERSION}
IMG_W_TAG = ${IMG}:${IMG_TAG}
DOCKERHUB_USERNAME ?= ""
DOCKERHUB_TOKEN ?= ""
GOOS ?= $(uname | tr '[:upper:]' '[:lower:]')
GOARCH ?= amd64
GOPROXY ?= "https://proxy.golang.org,direct"
MAKEFILE_PATH = $(dir $(realpath -s $(firstword $(MAKEFILE_LIST))))
BUILD_DIR_PATH = ${MAKEFILE_PATH}/build
SUPPORTED_PLATFORMS ?= "windows/amd64,darwin/amd64,linux/amd64,linux/arm64,linux/arm"
SELECTOR_PKG_VERSION_VAR=github.com/aws/amazon-ec2-instance-selector/pkg/selector.versionID
LATEST_RELEASE_TAG=$(shell git tag | tail -1)
PREVIOUS_RELEASE_TAG=$(shell git tag | tail -2 | head -1)

$(shell mkdir -p ${BUILD_DIR_PATH} && touch ${BUILD_DIR_PATH}/_go.mod)

repo-full-name:
	@echo ${REPO_FULL_NAME}

compile:
	go build -a -ldflags "-s -w -X main.versionID=${VERSION} -X ${SELECTOR_PKG_VERSION_VAR}=${VERSION}" -tags="aeis${GOOS}" -o ${BUILD_DIR_PATH}/${BIN} ${MAKEFILE_PATH}/cmd/main.go

clean:
	rm -rf ${BUILD_DIR_PATH}/ && go clean -testcache ./...

fmt:
	goimports -w ./ && gofmt -s -w ./

docker-build:
	${MAKEFILE_PATH}/scripts/build-docker-images -p ${GOOS}/${GOARCH} -r ${IMG} -v ${VERSION}

docker-run:
	docker run ${IMG_W_TAG}

docker-push:
	@echo ${DOCKERHUB_TOKEN} | docker login -u ${DOCKERHUB_USERNAME} --password-stdin
	docker push ${IMG_W_TAG}

build-docker-images:
	${MAKEFILE_PATH}/scripts/build-docker-images -p ${SUPPORTED_PLATFORMS} -r ${IMG} -v ${VERSION}

push-docker-images:
	@echo ${DOCKERHUB_TOKEN} | docker login -u ${DOCKERHUB_USERNAME} --password-stdin
	${MAKEFILE_PATH}/scripts/push-docker-images -p ${SUPPORTED_PLATFORMS} -r ${IMG} -v ${VERSION} -m

version:
	@echo ${VERSION}

latest-release-tag:
	@echo ${LATEST_RELEASE_TAG}

previous-release-tag:
	@echo ${PREVIOUS_RELEASE_TAG}

image:
	@echo ${IMG_W_TAG}

license-test:
	${MAKEFILE_PATH}/test/license-test/run-license-test.sh

go-report-card-test:
	${MAKEFILE_PATH}/test/go-report-card-test/run-report-card-test.sh

spellcheck:
	${MAKEFILE_PATH}/test/readme-test/run-readme-spellcheck

shellcheck:
	${MAKEFILE_PATH}/test/shellcheck/run-shellcheck

## requires aws credentials
readme-codeblock-test: 
	${MAKEFILE_PATH}/test/readme-test/run-readme-codeblocks

## requires aws credentials
output-validation-test: 
	${MAKEFILE_PATH}/test/output-validation-test/test-output-validation

build-binaries:
	${MAKEFILE_PATH}/scripts/build-binaries -p ${SUPPORTED_PLATFORMS} -v ${VERSION}

## requires a github token
upload-resources-to-github:
	${MAKEFILE_PATH}/scripts/upload-resources-to-github

## requires a dockerhub token
sync-readme-to-dockerhub:
	${MAKEFILE_PATH}/scripts/sync-readme-to-dockerhub

unit-test:
	go test -bench=. ${MAKEFILE_PATH}/...  -v -coverprofile=coverage.out -covermode=atomic -outputdir=${BUILD_DIR_PATH}

e2e-test: build
	${MAKEFILE_PATH}/test/e2e/run-test

homebrew-sync-dry-run:
	${MAKEFILE_PATH}/scripts/sync-to-aws-homebrew-tap -d -b ${BIN} -r ${REPO_FULL_NAME} -p ${SUPPORTED_PLATFORMS} -v ${LATEST_RELEASE_TAG}

homebrew-sync:
	${MAKEFILE_PATH}/scripts/sync-to-aws-homebrew-tap -b ${BIN} -r ${REPO_FULL_NAME} -p ${SUPPORTED_PLATFORMS}

build: compile

release: build-binaries build-docker-images push-docker-images upload-resources-to-github

test: spellcheck shellcheck unit-test license-test go-report-card-test e2e-test output-validation-test readme-codeblock-test

help:
	@grep -E '^[a-zA-Z_-]+:.*$$' $(MAKEFILE_LIST) | sort
