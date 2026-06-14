#!/usr/bin/env bash
set -euo pipefail

# bytemsg233 一键构建脚本
# Usage: bash scripts/build.sh [--version v1.0.0] [--os linux,darwin,windows] [--arch amd64,arm64]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

# Defaults
VERSION="${VERSION:-dev}"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo 'none')"
DATE="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
OS_LIST="linux,darwin,windows"
ARCH_LIST="amd64,arm64"
OUTPUT_DIR="dist"
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# Parse args
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)  VERSION="$2"; LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"; shift 2 ;;
        --os)       OS_LIST="$2"; shift 2 ;;
        --arch)     ARCH_LIST="$2"; shift 2 ;;
        --output)   OUTPUT_DIR="$2"; shift 2 ;;
        *)          echo "Unknown option: $1"; exit 1 ;;
    esac
done

log() { echo -e "${CYAN}[build]${NC} $*"; }
ok()  { echo -e "${GREEN}[BUILD OK]${NC} $*"; }

log "Version: ${VERSION}"
log "Commit:  ${COMMIT}"
log "Date:    ${DATE}"

rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

IFS=',' read -ra OS <<< "$OS_LIST"
IFS=',' read -ra ARCH <<< "$ARCH_LIST"

FAIL=0

for os in "${OS[@]}"; do
    for arch in "${ARCH[@]}"; do
        # Skip unsupported combos
        if [[ "$os" == "darwin" && "$arch" == "386" ]]; then continue; fi
        if [[ "$os" == "windows" && "$arch" == "arm64" ]]; then continue; fi

        EXT=""
        [[ "$os" == "windows" ]] && EXT=".exe"

        BIN_NAME="bytemsg233_${os}_${arch}${EXT}"
        log "Building ${BIN_NAME}..."

        GOOS="$os" GOARCH="$arch" go build \
            -ldflags="$LDFLAGS" \
            -trimpath \
            -o "${OUTPUT_DIR}/${BIN_NAME}" \
            ./cmd/bytemsg233

        if [[ $? -eq 0 ]]; then
            SIZE=$(du -h "${OUTPUT_DIR}/${BIN_NAME}" | cut -f1)
            ok "${BIN_NAME} (${SIZE})"
        else
            echo -e "${RED}FAILED${NC}: ${BIN_NAME}"
            FAIL=1
        fi
    done
done

# Generate checksums
log "Generating checksums..."
cd "$OUTPUT_DIR"
sha256sum * > checksums.txt 2>/dev/null || shasum -a 256 * > checksums.txt
cd "$PROJECT_ROOT"
ok "checksums.txt"

# Summary
echo ""
log "=== Build Summary ==="
ls -lh "$OUTPUT_DIR"/*.exe "$OUTPUT_DIR"/bytemsg233_* 2>/dev/null | awk '{print $5, $NF}'
echo ""

[[ $FAIL -eq 0 ]] && ok "All builds succeeded" || { echo -e "${RED}Some builds failed${NC}"; exit 1; }
