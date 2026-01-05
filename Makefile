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


COMMIT=$(shell git rev-parse --short HEAD)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
DIR=.rafal

provider:
	@go build -gcflags="all=-N -l" -o ~/.terraform.d/plugins/registry.terraform.io/gamefabric/gamefabric/1.0.0/darwin_arm64/terraform-provider-gamefabric_v1
.PHONY: provider

provider-local:
	@echo "==> Building provider from local code for $(GOOS)_$(GOARCH)"
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/gamefabric/gamefabric/1.0.0/$(GOOS)_$(GOARCH)
	@go build -gcflags="all=-N -l" -o ~/.terraform.d/plugins/registry.terraform.io/gamefabric/gamefabric/1.0.0/$(GOOS)_$(GOARCH)/terraform-provider-gamefabric_v1
	@echo "==> Provider built successfully at ~/.terraform.d/plugins/registry.terraform.io/gamefabric/gamefabric/1.0.0/$(GOOS)_$(GOARCH)/terraform-provider-gamefabric_v1"
.PHONY: provider-local

tf-init:
	@rm -rf $(DIR)/.terraform $(DIR)/.terraform.lock.hcl
	@terraform -chdir=$(DIR) init
.PHONY: tf-init

tf-apply:
	@terraform -chdir=$(DIR) apply
.PHONY: tf-apply

tf-apply-debug:
	@TF_LOG=DEBUG terraform -chdir=$(DIR) apply 2>&1 | tee terraform-apply.log
	@echo "==> Debug log saved to terraform-apply.log"
.PHONY: tf-apply-debug

tf-destroy:
	@terraform -chdir=$(DIR) destroy
.PHONY: tf-destroy