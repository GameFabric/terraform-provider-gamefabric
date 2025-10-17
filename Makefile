#Cleanup

fmt:
	@echo "==> Formatting source"
	@gofmt -s -w $(shell find . -type f -name '*.go' -not -path "./vendor/*")
	@echo "==> Done"
.PHONY: fmt

#Test

test:
	@go test -cover -race ./...
.PHONY: test

#Lint

lint:
	@golangci-lint run --config=.golangci.yml ./...
.PHONY: lint

#Build

build:
	@goreleaser release --clean --snapshot --skip=sign
.PHONY: build

# Docs Generation

docs-gen:
	@go tool tfplugindocs generate  --rendered-provider-name="GameFabric"
.PHONY: docs-gen
