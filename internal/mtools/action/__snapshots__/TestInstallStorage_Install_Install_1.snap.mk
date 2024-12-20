####################################################################################################
## DB COMMANDS
####################################################################################################
.PHONY: db
db: ## Run all db commands
    go install github.com/go-modulus/modulus/cmd/mtools@latest
    $(MAKE) db-update-sqlc
    $(MAKE) db-migrate

.PHONY: db-migrate
migrate: ## Run migrations from all modules
    go run cmd/console/main.go migrator migrate

.PHONY: db-update-sqlc
db-update-sqlc: ## Update sqlc.yaml configs in all modules combining definitions from the sqlc.definition.yaml and sqlc.tmpl.yaml
    mtools db update-sqlc-config

####################################################################################################
## END OF DB COMMANDS
####################################################################################################
