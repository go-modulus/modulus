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
	go run github.com/rakyll/gotest -v -failfast  ./...

analyze: ## Run static analyzer
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.61.0 golangci-lint run -v

install: ## Make a binary to ./bin folder
	go build -o ./bin/mtools  ./cmd/mtools/main.go

update-manifest: ## Update the manifest file
	go run ./cmd/manifest/main.go

build-testproject: ## Build the example of a project
	$(MAKE) install
	./bin/mtools init --path=./testproj --name=testproj
	./bin/mtools module install --proj-path=./testproj -m "pgx"
	./bin/mtools module create --proj-path=./testproj --silent --path=internal --package=example
	./bin/mtools db add --proj-path=./testproj --module=example --name=create_table
	echo "-- migrate:up" > ./testproj/internal/example/storage/migration/20241228085104_create_table.sql
	echo "CREATE TABLE example (" >> ./testproj/internal/example/storage/migration/20241228085104_create_table.sql
	echo "	id SERIAL PRIMARY KEY," >> ./testproj/internal/example/storage/migration/20241228085104_create_table.sql
	echo "	name TEXT NOT NULL" >> ./testproj/internal/example/storage/migration/20241228085104_create_table.sql
	echo ");" >> ./testproj/internal/example/storage/migration/20241228085104_create_table.sql
	echo "-- migrate:down" >> ./testproj/internal/example/storage/migration/20241228085104_create_table.sql
	echo "DROP TABLE example;" >> ./testproj/internal/example/storage/migration/20241228085104_create_table.sql
	echo "-- name: FindExamples :many" > ./testproj/internal/example/storage/query/example.sql
	echo "SELECT * FROM example;" >> ./testproj/internal/example/storage/query/example.sql
	./bin/mtools db update-sqlc-config --proj-path=./testproj
	./bin/mtools db generate --proj-path=./testproj
	./bin/mtools db migrate --proj-path=./testproj
	./bin/mtools module install --proj-path=./testproj -m "dbmate migrator"
	cd ./testproj && go run cmd/console/main.go migrator migrate
	./bin/mtools module add-cli --proj-path=./testproj --module=example --name=hello-world
	cd ./testproj && go run cmd/console/main.go hello-world
	./bin/mtools module install --proj-path=./testproj --manifest="modules.json" -m "chi http"
	./bin/mtools module add-json-api --proj-path=./testproj --module=example --uri=/hello-world --name=HelloWorld --method=GET --silent
	./bin/mtools module install --proj-path=./testproj --manifest="modules.json" -m "gqlgen"

.PHONY: db-sqlc-generate
db-sqlc-generate: ## Generate sqlc files in all modules
	sqlc -f auth/storage/sqlc.yaml generate

.PHONY: db-migrate
db-migrate: ## Run migrations in test database
	$(MAKE) install
	APP_ENV=test ./bin/mtools db migrate --local-manifest=modules-test.json