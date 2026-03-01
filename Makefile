default: help

###
## Add these lines to your .zshrc to have autocompletion for make commands
## zstyle ':completion:*:make:*:targets' call-command true
## zstyle ':completion:*:*:make:*' tag-order 'targets'
##
####################################################################################################
## MAIN COMMANDS
####################################################################################################
.PHONY: help
help: ## show this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: test
test: ## Run tests
	go run github.com/rakyll/gotest@latest -v -failfast  ./...

analyze: ## Run static analyzer
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v2.10.1 golangci-lint run -v

.PHONY: db-sqlc-generate
db-sqlc-generate: ## Generate sqlc files in all modules
	sqlc -f auth/storage/sqlc.yaml generate

.PHONY: db-migrate
db-migrate: ## Run migrations in test database
	$(MAKE) install
	APP_ENV=test ./bin/mtools db migrate --local-manifest=modules-test.json

.PHONY: db-migrate
db-rollback: ## Rollback the last migration in test database
	$(MAKE) install
	APP_ENV=test ./bin/mtools db rollback --local-manifest=modules-test.json

.PHONY: translation-extract
translation-extract: ## Extract translations from source code
	@echo "Extracting translations..."
	go run github.com/go-modulus/xspreak@latest -D ./auth -p ./auth/locales -d auth
