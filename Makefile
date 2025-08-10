DEV_BIN:=$(shell pwd)/dev_tools/bin
MAKEFILE_DIR := $(shell cd $(dir $(lastword $(MAKEFILE_LIST)))&&pwd )

.PHONY: build
build:
	go build -o hatsukari ./...

.PHONY: fmt
fmt: $(DEV_BIN)/golangci-lint
	$(DEV_BIN)/golangci-lint run --fix --config=.golangci.yml

.PHONY: lint
lint: $(DEV_BIN)/golangci-lint
	$(DEV_BIN)/golangci-lint run --config=.golangci.yml

.PHONY: setup
setup: $(DEV_BIN)/air $(DEV_BIN)/dlv $(DEV_BIN)/golangci-lint

$(DEV_BIN)/air:
	mkdir -p $(@D)
	GOBIN=$(DEV_BIN) go install github.com/air-verse/air@v1.62.0


$(DEV_BIN)/golangci-lint:
	mkdir -p $(@D)
	GOBIN=$(DEV_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.3.1

$(DEV_BIN)/dlv:
	mkdir -p $(@D)
	GOBIN=$(DEV_BIN) go install github.com/go-delve/delve/cmd/dlv@latest

.PHONY: ci
ci: lint

.PHONY: clean
clean:
	rm -rf dev_tools/bin


