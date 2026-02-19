.PHONY: help build install test clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build vlt binary
	go build -o vlt .

install: ## Install vlt to $GOPATH/bin
	go install .

test: ## Run tests
	go test -v ./...

clean: ## Remove build artifacts
	rm -f vlt
