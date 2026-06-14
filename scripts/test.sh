#!/usr/bin/env bash
set -euo pipefail

# bytemsg233 一键测试脚本
# Usage: bash scripts/test.sh [--coverage] [--bench] [--race] [--verbose]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# Defaults
COVERAGE=false
BENCH=false
RACE=false
VERBOSE=false
PACKAGES="./..."

# Parse args
while [[ $# -gt 0 ]]; do
    case $1 in
        --coverage) COVERAGE=true; shift ;;
        --bench)    BENCH=true; shift ;;
        --race)     RACE=true; shift ;;
        --verbose)  VERBOSE=true; shift ;;
        --pkg)      PACKAGES="$2"; shift 2 ;;
        *)          echo "Unknown option: $1"; exit 1 ;;
    esac
done

log() { echo -e "${CYAN}[test]${NC} $*"; }
ok()  { echo -e "${GREEN}[PASS]${NC} $*"; }
err() { echo -e "${RED}[FAIL]${NC} $*"; }

# Clean previous coverage
rm -f coverage.out coverage.html

# Build test flags
TEST_FLAGS=""
[[ "$VERBOSE" == true ]] && TEST_FLAGS="$TEST_FLAGS -v"
[[ "$RACE" == true ]] && TEST_FLAGS="$TEST_FLAGS -race"
[[ "$COVERAGE" == true ]] && TEST_FLAGS="$TEST_FLAGS -coverprofile=coverage.out -covermode=atomic"

# Run tests
log "Running tests..."
if go test $TEST_FLAGS $PACKAGES; then
    ok "All tests passed"
else
    err "Tests failed"
    exit 1
fi

# Coverage report
if [[ "$COVERAGE" == true ]]; then
    log "Generating coverage report..."
    go tool cover -func=coverage.out | tail -1
    go tool cover -html=coverage.out -o coverage.html 2>/dev/null && \
        log "HTML report: coverage.html" || true

    # Check threshold
    TOTAL=$(go tool cover -func=coverage.out | tail -1 | grep -oP '[\d.]+(?=%)' || echo "0")
    if (( $(echo "$TOTAL < 80" | bc -l) )); then
        echo -e "${YELLOW}WARNING: Coverage ${TOTAL}% is below 80% threshold${NC}"
    else
        ok "Coverage ${TOTAL}% meets 80% threshold"
    fi
fi

# Benchmarks
if [[ "$BENCH" == true ]]; then
    log "Running benchmarks..."
    go test ./pkg/binary/... -bench=. -benchmem -count=3
fi

# Size comparison
log "Running size comparison..."
go test ./pkg/binary/... -run "TestSizeComparison" -v

log "Done."
