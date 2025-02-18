include .env.local

.DEFAULT_GOAL := help

export $(shell sed 's/=.*//' .env.local)

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

dev: ## Run air for live development
	GO_ENV=development air & pnpm --dir frontend dev --host

local-release: ## Build the project locally, skipping signing
	goreleaser release --snapshot --clean --skip sign

test: ## Run tests
	go test ./...

generate: ## Generate code
	go generate ./...

fmt: ## Format the code using gofmt
	gofmt -s -w .

up-db: ## Start the local dependencies stack
	docker compose up -d

down-db: ## Stop the local dependency stack
	docker compose down

migration: ## Create a new SQL migration
	sql-migrate new $(name)
