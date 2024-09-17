DEFAULT_GOAL := help
RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(eval $(RUN_ARGS):;@:)

define check_console
	test -s ./bin/console ||(go build -o ./bin/console  ./cmd/console/main.go)
endef

####################################################################################################
## MAIN COMMANDS
####################################################################################################
help: ## Commands list
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'


install: ## Make a binary to ./bin folder
	go build -o ./bin/console  ./cmd/console/main.go

analyze: ## Run static analyzer
	test -s ./bin/golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.59.1
	./bin/golangci-lint run -c ./.golangci.yaml ./...

run: ## Run server
	$(MAKE) ./bin/console serve

test: ## Run tests
	test -s ./bin/gotest || go build -o ./bin/gotest github.com/rakyll/gotest
	./bin/gotest -v -failfast  ./internal/...

cover: ## Run tests with coverage
	go test -v -failfast -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: mocks
mocks:
	go run github.com/vektra/mockery/v2/

####################################################################################################
## END OF MAIN COMMANDS
####################################################################################################
