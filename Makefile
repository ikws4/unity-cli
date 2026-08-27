GO ?= go
BINARY ?= unity-cli
BUILD_DIR ?= bin
GOBIN ?= $(shell $(GO) env GOBIN)
ifeq ($(strip $(GOBIN)),)
GOBIN := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))/bin
endif

.PHONY: build install test fmt

build:
	mkdir -p "$(BUILD_DIR)"
	$(GO) build -trimpath -o "$(BUILD_DIR)/$(BINARY)" .

install:
	GOBIN="$(GOBIN)" $(GO) install -trimpath .
	@echo "installed $(GOBIN)/$(BINARY)"

test:
	$(GO) test ./...

fmt:
	gofmt -w .
