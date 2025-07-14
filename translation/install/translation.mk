.PHONY: translation-extract
translation-extract: ## Extract translations from source code
	@echo "Extracting translations..."
	go run github.com/vorlif/xspreak@latest -D ./internal -p ./locales