SWEEP?=us-east-1,us-west-2
TEST?=./...
SWEEP_DIR?=./victorops
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=victorops
WEBSITE_REPO=github.com/hashicorp/terraform-website
TEST_COUNT?=1

default: build

build: fmtcheck
	go build -o terraform-provider-victorops

install: build
	go install

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test $(SWEEP_DIR) -v -sweep=$(SWEEP) $(SWEEPARGS) -timeout 60m

test: fmtcheck
	go test $(TEST) $(TESTARGS) -short -timeout=120s -parallel=4

testacc: fmtcheck
	@echo "==> Running acceptance tests..."
	@echo "Note: Set VO_API_ID, VO_API_KEY, VO_BASE_URL, and VO_REPLACEMENT_USERNAME for acceptance tests"
	go test $(TEST) -v -count $(TEST_COUNT) -parallel 20 $(TESTARGS) -timeout 120m

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w ./$(PKG_NAME)

fmtcheck:
	@echo "==> Checking source code formatting..."
	@test -z "$$(gofmt -s -l ./$(PKG_NAME) | tee /dev/stderr)" || \
		(echo; echo "Please run 'make fmt' to fix formatting"; exit 1)

vet:
	@echo "==> Running go vet..."
	@go vet ./...

depscheck:
	@echo "==> Checking source code with go mod tidy..."
	@go mod tidy
	@git diff --exit-code -- go.mod go.sum || \
		(echo; echo "Unexpected difference in go.mod/go.sum files. Run 'go mod tidy' command or revert any go.mod/go.sum changes and commit."; exit 1)

lint: fmtcheck vet
	@echo "==> Checking source code against linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./$(PKG_NAME)/...; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

docs:
	@echo "==> Generating documentation..."
	@if command -v tfplugindocs >/dev/null 2>&1; then \
		tfplugindocs generate; \
	else \
		echo "tfplugindocs not installed. Install with: go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest"; \
	fi

tools:
	@echo "==> Installing development tools..."
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

clean:
	rm -f terraform-provider-victorops

.PHONY: build install sweep test testacc fmt fmtcheck vet lint tools test-compile depscheck docs clean
