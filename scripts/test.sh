#!/usr/bin/env bash
# scripts/test.sh — Run all Go tests with coverage reporting
# Usage: ./scripts/test.sh [--coverage] [--json]
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

BACKEND_DIR="backend"
PHOENIX_DIR="phoenix"
COVERAGE_DIR="coverage"
DO_COVERAGE=false
JSON_OUTPUT=false

for arg in "$@"; do
  case $arg in
    --coverage) DO_COVERAGE=true ;;
    --json)     JSON_OUTPUT=true ;;
  esac
done

TOTAL_PASS=0
TOTAL_FAIL=0
TOTAL_SKIP=0
START_TIME=$(date +%s)

echo ""
echo "═══════════════════════════════════════════════"
echo "  Challenge To YOU — Test Runner"
echo "═══════════════════════════════════════════════"
echo ""

for MODULE in "$BACKEND_DIR" "$PHOENIX_DIR"; do
  if [ ! -d "$MODULE" ]; then
    echo -e "${YELLOW}SKIP${NC}: $MODULE not found"
    TOTAL_SKIP=$((TOTAL_SKIP + 1))
    continue
  fi

  echo "── Module: $MODULE ──"

  TEST_ARGS="-count=1 -timeout 60s"

  if $DO_COVERAGE; then
    mkdir -p "$COVERAGE_DIR"
    TEST_ARGS="$TEST_ARGS -coverprofile=$COVERAGE_DIR/${MODULE//\//_}.out -covermode=atomic"
  fi

  START=$(date +%s%N)
  OUTPUT=$(cd "$MODULE" && go test $TEST_ARGS ./... 2>&1) && TEST_EXIT=0 || TEST_EXIT=$?
  END=$(date +%s%N)
  DURATION_MS=$(( (END - START) / 1000000 ))

  PASSED=$(echo "$OUTPUT" | grep -c "^ok" || true)
  FAILED=$(echo "$OUTPUT" | grep -c "^FAIL" || true)

  TOTAL_PASS=$((TOTAL_PASS + PASSED))
  TOTAL_FAIL=$((TOTAL_FAIL + FAILED))

  if [ $TEST_EXIT -ne 0 ]; then
    echo -e "${RED}FAIL${NC} ($DURATION_MS ms)"
    echo "$OUTPUT"
  else
    echo -e "${GREEN}PASS${NC} ($DURATION_MS ms) — $PASSED package(s)"
  fi

  if $DO_COVERAGE && [ -f "$COVERAGE_DIR/${MODULE//\//_}.out" ]; then
    echo "  Coverage: $(go tool cover -func="$COVERAGE_DIR/${MODULE//\//_}.out" | tail -1)"
  fi

  echo ""
done

END_TIME=$(date +%s)
TOTAL_DURATION=$((END_TIME - START_TIME))

echo "═══════════════════════════════════════════════"
echo "  Summary"
echo "═══════════════════════════════════════════════"
echo -e "  Passed:  ${GREEN}$TOTAL_PASS${NC}"
echo -e "  Failed:  ${RED}$TOTAL_FAIL${NC}"
echo -e "  Skipped: ${YELLOW}$TOTAL_SKIP${NC}"
echo "  Duration: ${TOTAL_DURATION}s"
echo ""

if $JSON_OUTPUT; then
  cat <<EOF
{"passed":$TOTAL_PASS,"failed":$TOTAL_FAIL,"skipped":$TOTAL_SKIP,"duration_s":$TOTAL_DURATION}
EOF
fi

if [ $TOTAL_FAIL -gt 0 ]; then
  exit 1
fi
