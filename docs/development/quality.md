# Quality Pipeline

Automated quality checks for Challenge To YOU. This document explains how to use the local tooling, CI, and pre-commit hooks.

---

## Quick Start

```bash
# Run all quality checks at once
make verify

# Or individually:
make test        # run all Go tests
make lint        # run golangci-lint
make vet         # run go vet
make fmt         # format all Go files
make fmtcheck    # check formatting (CI-friendly, no modifications)
make build       # compile backend binary
make coverage    # generate coverage report with threshold enforcement
```

---

## Prerequisites

| Tool | Install | Purpose |
|------|---------|---------|
| Go 1.25+ | [go.dev](https://go.dev/dl/) | Language runtime |
| golangci-lint | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` | Linting |
| goimports | `go install golang.org/x/tools/cmd/goimports@latest` | Import formatting |

---

## Local Workflow

### Before committing

```bash
# 1. Format your code
make fmt

# 2. Run all checks
make verify

# 3. If everything passes, commit
git add -A && git commit -m "feat: ..."
```

### Install pre-commit hook (one-time)

```bash
make install-hooks
```

The hook automatically runs:
- `gofmt` formatting check
- `go vet` on changed packages
- `go test` on changed test files

Commits are blocked if any check fails.

---

## CI (GitHub Actions)

Every push and PR triggers `.github/workflows/quality.yml`:

| Job | What it checks |
|-----|----------------|
| **fmt** | `gofmt -l` — no unformatted files |
| **vet** | `go vet ./...` on both modules |
| **lint** | `golangci-lint run` on `backend/` and `phoenix/` |
| **test** | `go test` with coverage on both modules |
| **build** | Compiles `backend/sandbox` and `phoenix/phoenix` |
| **markdown** | Link checking across all `.md` files |
| **json-schema** | Validates challenge JSON against schemas in `content/schema/` |

The **build** job depends on all others passing first.

---

## Coverage

### Generate a report

```bash
make coverage
```

This produces:
- `coverage/coverage.out` — raw coverage data
- `coverage/coverage.html` — browser-viewable report

### Threshold

Coverage is enforced at **40%** by default. If coverage drops below this, `make coverage` and CI fail.

To change the threshold:

```bash
make coverage COVERAGE_THRESH=50
```

Or in CI, set the `COVERAGE_THRESH` environment variable in the workflow.

### Interpreting output

```
── Coverage Summary ──
total: (statements) 45.2%
Threshold: 40%
```

---

## Linting Configuration

The linter config lives at the repository root: `.golangci.yml`

**Enabled linters:**
- `errcheck` — unhandled errors
- `gosimple` — simplification suggestions
- `govet` — suspicious constructs
- `staticcheck` — comprehensive static analysis
- `unused` — dead code
- `gofmt` — formatting
- `goimports` — import ordering
- `misspell` — typos
- `unconvert` — unnecessary conversions
- `gocritic` — opinionated style checks
- `gosec` — security issues

**Exclusions:**
- `gosec` and `errcheck` are disabled in `_test.go` files

---

## Troubleshooting

### "golangci-lint not installed"

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### "goimports not found"

```bash
go install golang.org/x/tools/cmd/goimports@latest
```

### Formatting fails in CI but passes locally

Ensure you're running `gofmt -s` (simplification flag). The Makefile and CI both use `-s`.

### Coverage threshold too strict

Lower the threshold: `make coverage COVERAGE_THRESH=20`

### Pre-commit hook blocks my commit

Run `make fmt` first to auto-fix formatting, then retry the commit.

---

## File Reference

| File | Purpose |
|------|---------|
| `Makefile` | All quality targets |
| `.golangci.yml` | Linter configuration |
| `.github/workflows/quality.yml` | CI pipeline |
| `.markdownlint.json` | Markdown linting config |
| `scripts/test.sh` | Standalone test runner |
| `scripts/pre-commit` | Git pre-commit hook |
| `coverage/` | Generated coverage reports (gitignored) |

---

*Last updated: 2026-07-11*
