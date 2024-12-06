.DEFAULT_GOAL := help

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

dev: ## Run air for live development
	air

local-release: ## Build the project locally, skipping signing
	goreleaser release --snapshot --clean --skip sign

test: ## Run tests
	go test ./...

generate: ## Generate code
	go generate ./...

fmt: ## Format the code using gofmt
	gofmt -s -w .
