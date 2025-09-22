.PHONY: translation-extract
translation-extract: ## Extract translations from source code
	@echo "Extracting translations..."
	go run github.com/go-modulus/xspreak@v0.1.0 -D ./internal -p ./locales --loaded-packages=github.com/go-modulus/modulus/errors