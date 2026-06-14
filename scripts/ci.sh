#!/usr/bin/env bash
set -euo pipefail

# bytemsg233 CI 全流程脚本
# Usage: bash scripts/ci.sh [--version v1.0.0]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

VERSION="${1:-dev}"

log() { echo -e "${CYAN}[ci]${NC} $*"; }
ok()  { echo -e "${GREEN}[CI OK]${NC} $*"; }
fail() { echo -e "${RED}[CI FAIL]${NC} $*"; exit 1; }

log "=== bytemsg233 CI Pipeline ==="
log "Version: $VERSION"
echo ""

# 1. Lint
log "Step 1/6: Lint..."
if command -v golangci-lint &>/dev/null; then
    golangci-lint run ./... || fail "Lint failed"
    ok "Lint passed"
else
    echo "  golangci-lint not found, running go vet instead"
    go vet ./... || fail "Vet failed"
    ok "Vet passed"
fi

# 2. Test
log "Step 2/6: Test..."
go test ./... -race -count=1 || fail "Tests failed"
ok "Tests passed"

# 3. Coverage
log "Step 3/6: Coverage..."
go test ./... -coverprofile=coverage.out -covermode=atomic 2>/dev/null
TOTAL=$(go tool cover -func=coverage.out | tail -1 | grep -oP '[\d.]+(?=%)' || echo "0")
log "Coverage: ${TOTAL}%"
if (( $(echo "$TOTAL < 70" | bc -l 2>/dev/null || echo 0) )); then
    echo "  WARNING: Coverage below 70%"
fi

# 4. Size comparison
log "Step 4/6: Size comparison..."
go test ./pkg/binary/... -run "TestSizeComparison" -v 2>&1 | grep -E "(bytes|节省|ByteMsg|Protobuf|MsgPack|JSON)" || true

# 5. Build all platforms
log "Step 5/6: Cross-platform build..."
bash scripts/build.sh --version "$VERSION" || fail "Build failed"

# 6. Verify CLI
log "Step 6/6: CLI verification..."
if [[ -f "dist/bytemsg233_linux_amd64" ]]; then
    # Can't run linux binary on macOS/Windows, but we can check it exists
    ok "Linux binary exists"
elif [[ -f "dist/bytemsg233.exe" ]]; then
    ./dist/bytemsg233.exe version
    ok "CLI works"
elif [[ -f "dist/bytemsg233_darwin_amd64" ]]; then
    ./dist/bytemsg233_darwin_amd64 version
    ok "CLI works"
fi

echo ""
ok "=== CI Pipeline Complete ==="
echo ""
log "Artifacts:"
ls -lh dist/ 2>/dev/null | tail -n +2 | awk '{print "  " $5 " " $NF}'
