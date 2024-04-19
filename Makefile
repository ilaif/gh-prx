.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

ide-setup: ## Installs specific requirements for local development
	go mod tidy
	pre-commit install

lint: ## Run lint
	golangci-lint run --fix ./...

test: ## Run unit tests
	go test -short ./...

testwatch: ## Run unit tests in watch mode, re-running tests on each file change
	-gotestsum --format pkgname -- -short ./...
	gotestsum --watch --format pkgname -- -short ./...

build: ## Build the binary
	go build ./

install-extension-local: build ## Installs the extension locally
	gh extension remove prx || true
	gh extension install .

install-extension-remote: ## Installs the extension remotely
	gh extension remove prx || true
	gh extension install ilaif/gh-prx
