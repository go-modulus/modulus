.PHONY: temporal-run-worker
temporal-run-worker: ## Run Default Temporal worker for the default queue
	$(MAKE) install
	./bin/console temporal worker -q default