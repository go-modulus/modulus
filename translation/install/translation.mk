.PHONY: translation-extract
translation-extract: ## Extract translations from source code
	@echo "Extracting translations..."
	go run github.com/vorlif/xspreak@latest -D ./internal -p ./locales --loaded-packages=github.com/go-modulus/modulus/errors

.PHONY: translation-compile
translation-compile: ## Compile translation files
	@echo "Compiling translations..."
	msgfmt -o messages.mo messages.po