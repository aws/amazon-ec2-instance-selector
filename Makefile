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
SUPPORTED_PLATFORMS ?= "windows/amd64,darwin/amd64,darwin/arm64,linux/amd64,linux/arm64,linux/arm"
SELECTOR_PKG_VERSION_VAR=github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector.versionID
LATEST_RELEASE_TAG=$(shell git describe --tags --abbrev=0)
PREVIOUS_RELEASE_TAG=$(shell git describe --abbrev=0 --tags `git rev-list --tags --skip=1  --max-count=1`)

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
	@docker login -u ${DOCKERHUB_USERNAME} -p="${DOCKERHUB_TOKEN}"
	docker push ${IMG_W_TAG}

build-docker-images:
	${MAKEFILE_PATH}/scripts/build-docker-images -p ${SUPPORTED_PLATFORMS} -r ${IMG} -v ${VERSION}

push-docker-images:
	@docker login -u ${DOCKERHUB_USERNAME} -p="${DOCKERHUB_TOKEN}"
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

spellcheck:
	${MAKEFILE_PATH}/test/readme-test/run-readme-spellcheck

shellcheck:
	${MAKEFILE_PATH}/test/shellcheck/run-shellcheck

## requires aws credentials
readme-codeblock-test: 
	${MAKEFILE_PATH}/test/readme-test/run-readme-codeblocks


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

## requires aws credentials
e2e-test: build
	${MAKEFILE_PATH}/test/e2e/run-test

## requires aws credentials
integ-test: e2e-test readme-codeblock-test

homebrew-sync-dry-run:
	${MAKEFILE_PATH}/scripts/sync-to-aws-homebrew-tap -d -b ${BIN} -r ${REPO_FULL_NAME} -p ${SUPPORTED_PLATFORMS} -v ${LATEST_RELEASE_TAG}

homebrew-sync:
	${MAKEFILE_PATH}/scripts/sync-to-aws-homebrew-tap -b ${BIN} -r ${REPO_FULL_NAME} -p ${SUPPORTED_PLATFORMS}

build: compile

release: build-binaries upload-resources-to-github

test: spellcheck shellcheck unit-test license-test e2e-test readme-codeblock-test

help:
	@grep -E '^[a-zA-Z_-]+:.*$$' $(MAKEFILE_LIST) | sort

## Targets intended to be run in preparation for a new release
create-local-release-tag-major:
	${MAKEFILE_PATH}/scripts/create-local-tag-for-release -m

create-local-release-tag-minor:
	${MAKEFILE_PATH}/scripts/create-local-tag-for-release -i

create-local-release-tag-patch:
	${MAKEFILE_PATH}/scripts/create-local-tag-for-release -p

create-release-prep-pr:
	${MAKEFILE_PATH}/scripts/prepare-for-release

create-release-prep-pr-draft:
	${MAKEFILE_PATH}/scripts/prepare-for-release -d

release-prep-major: create-local-release-tag-major create-release-prep-pr

release-prep-minor: create-local-release-tag-minor create-release-prep-pr

release-prep-patch: create-local-release-tag-patch create-release-prep-pr

release-prep-custom: # Run make NEW_VERSION=v1.2.3 release-prep-custom to prep for a custom release version
ifdef NEW_VERSION
	$(shell echo "${MAKEFILE_PATH}/scripts/create-local-tag-for-release -v $(NEW_VERSION) && echo && make create-release-prep-pr")
endif
