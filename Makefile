.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

ide-setup: ## Installs specific requirements for local development
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0
	go install gotest.tools/gotestsum@v1.8.2
	pre-commit install

lint: ## Run lint
	golangci-lint run ./...

test: ## Run unit tests
	go test -short ./...

testwatch: ## Run unit tests in watch mode, re-running tests on each file change
	-gotestsum --format pkgname -- -short ./...
	gotestsum --watch --format pkgname -- -short ./...

build: ## Build the binary
	go build ./

install-extension-local: build ## Installs the extension locally
	gh extension remove prx
	gh extension install .

install-extension-remote: ## Installs the extension remotely
	gh extension remove prx
	gh extension install prx
