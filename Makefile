.PHONY: test lint build fmt fmtcheck vet verify coverage clean install-hooks

# ── Go modules ──────────────────────────────────────────────
BACKEND_DIR  := backend
PHOENIX_DIR  := phoenix
GO_MODULES   := $(BACKEND_DIR) $(PHOENIX_DIR)

# ── Coverage ────────────────────────────────────────────────
COVERAGE_DIR    := coverage
COVERAGE_FILE   := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML   := $(COVERAGE_DIR)/coverage.html
COVERAGE_THRESH ?= 40

# ── Tools ───────────────────────────────────────────────────
GOBIN          := $(shell go env GOPATH)/bin
GOLANGCI_LINT  := $(GOBIN)/golangci-lint
GOIMPORTS      := $(GOBIN)/goimports

# ═══════════════════════════════════════════════════════════
#  TARGETS
# ═══════════════════════════════════════════════════════════

## test: Run all Go tests across every module
test:
	@echo "==> Running tests..."
	@EXIT_CODE=0; \
	for dir in $(GO_MODULES); do \
		echo "--- $$dir ---"; \
		(cd $$dir && go test -count=1 -timeout 60s ./...) || EXIT_CODE=$$?; \
	done; \
	if [ $$EXIT_CODE -ne 0 ]; then \
		echo "==> Tests FAILED."; \
		exit $$EXIT_CODE; \
	fi
	@echo "==> All tests passed."

## lint: Run golangci-lint on backend (primary module)
lint:
	@if [ ! -f "$(GOLANGCI_LINT)" ]; then \
		echo "ERROR: golangci-lint not installed. Run:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	@echo "==> Linting backend..."
	cd $(BACKEND_DIR) && $(GOLANGCI_LINT) run ./...
	@echo "==> Linting phoenix..."
	cd $(PHOENIX_DIR) && $(GOLANGCI_LINT) run ./...
	@echo "==> Lint passed."

## vet: Run go vet on all modules
vet:
	@echo "==> Running go vet..."
	@EXIT_CODE=0; \
	for dir in $(GO_MODULES); do \
		echo "--- $$dir ---"; \
		(cd $$dir && go vet ./...) || EXIT_CODE=$$?; \
	done; \
	if [ $$EXIT_CODE -ne 0 ]; then echo "==> Vet FAILED."; exit $$EXIT_CODE; fi
	@echo "==> Vet passed."

## fmt: Format all Go source files
fmt:
	@echo "==> Formatting..."
	@for dir in $(GO_MODULES); do \
		(echo "--- $$dir ---" && cd $$dir && gofmt -s -w . && $(GOIMPORTS) -w .); \
	done
	@echo "==> Formatting done."

## fmtcheck: Verify formatting without modifying files (CI-friendly)
fmtcheck:
	@echo "==> Checking formatting..."
	@EXIT_CODE=0; \
	for dir in $(GO_MODULES); do \
		UNCHANGED=$$(cd $$dir && gofmt -l .); \
		if [ -n "$$UNCHANGED" ]; then \
			echo "Files need formatting in $$dir:"; \
			echo "$$UNCHANGED"; \
			EXIT_CODE=1; \
		fi; \
	done; \
	exit $$EXIT_CODE
	@echo "==> Formatting OK."

## build: Compile the backend binary
build:
	@echo "==> Building backend..."
	cd $(BACKEND_DIR) && go build -o sandbox ./cmd/sandbox/
	@echo "==> Build complete: backend/sandbox"

## coverage: Generate coverage report and enforce threshold
coverage:
	@mkdir -p $(COVERAGE_DIR)
	@echo "==> Generating coverage..."
	cd $(BACKEND_DIR) && go test -coverprofile=../$(COVERAGE_FILE) -covermode=atomic ./...
	@echo ""
	@echo "── Coverage Summary ──"
	@go tool cover -func=$(COVERAGE_FILE) | tail -1
	@TOTAL=$$(go tool cover -func=$(COVERAGE_FILE) | tail -1 | awk '{print $$NF}' | tr -d '%'); \
	echo "Threshold: $(COVERAGE_THRESH)%"; \
	if [ "$$(echo "$$TOTAL < $(COVERAGE_THRESH)" | bc -l)" = "1" ]; then \
		echo "FAIL: Coverage $$TOTAL% is below threshold $(COVERAGE_THRESH)%"; \
		exit 1; \
	fi
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "==> HTML report: $(COVERAGE_HTML)"

## verify: Run all quality checks (fmtcheck + vet + lint + test)
verify: fmtcheck vet lint test
	@echo ""
	@echo "══════════════════════════════════════"
	@echo "  ALL QUALITY CHECKS PASSED"
	@echo "══════════════════════════════════════"

## install-hooks: Install git pre-commit hook
install-hooks:
	@echo "==> Installing pre-commit hook..."
	@mkdir -p .git/hooks
	@cp scripts/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "==> Pre-commit hook installed."

## clean: Remove build artifacts and coverage
clean:
	@echo "==> Cleaning..."
	rm -f $(BACKEND_DIR)/sandbox
	rm -rf $(COVERAGE_DIR)
	@echo "==> Clean."
