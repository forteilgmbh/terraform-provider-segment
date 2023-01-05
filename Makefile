.EXPORT_ALL_VARIABLES:
VERSION := 0.4.3
NAME := terraform-provider-segment_v${VERSION}
BUILD_DIR := /usr/local/share/terraform/plugins/terraform.bonify.de/forteilgmbh/segment/${VERSION}/linux_amd64
TARGET := ${BUILD_DIR}/${NAME}
LDFLAGS ?=

.PHONY: build
build: build-linux ## run go build for linux

.PHONY: build-native
build-native: ## run go build for current OS
	@go build -ldflags "$(LDFLAGS)" -o "${TARGET}"

.PHONY: build-linux
build-linux:
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o "${TARGET}"

.PHONY: test
test: ## Runs the go tests.
	@go test -v ./...

.PHONY: testacc
testacc:
	@TF_ACC=1 go test -v ./...

.PHONY: release
release:
	@goreleaser --rm-dist
