analyze: ## Run static analyzer
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.61.0 golangci-lint run -v

install: ## Make a binary to ./bin folder
	go build -o ./bin/mtools  ./cmd/mtools/main.go