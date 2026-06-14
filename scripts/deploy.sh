#!/usr/bin/env bash
set -euo pipefail

# bytemsg233 发布脚本
# Usage: bash scripts/deploy.sh [--version v1.0.0] [--dry-run] [--skip-test]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

VERSION=""
DRY_RUN=false
SKIP_TEST=false
SKIP_BUILD=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --version)    VERSION="$2"; shift 2 ;;
        --dry-run)    DRY_RUN=true; shift ;;
        --skip-test)  SKIP_TEST=true; shift ;;
        --skip-build) SKIP_BUILD=true; shift ;;
        *)            echo "Unknown option: $1"; exit 1 ;;
    esac
done

log()  { echo -e "${CYAN}[deploy]${NC} $*"; }
ok()   { echo -e "${GREEN}[DEPLOY OK]${NC} $*"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
err()  { echo -e "${RED}[ERR]${NC} $*"; }

# Validate git state
if [[ -n "$(git status --porcelain)" ]]; then
    warn "Working tree has uncommitted changes"
    git status --short
    echo ""
fi

CURRENT_BRANCH="$(git branch --show-current)"
log "Current branch: $CURRENT_BRANCH"

# Determine version
if [[ -z "$VERSION" ]]; then
    VERSION="$(git describe --tags --always 2>/dev/null || echo 'dev')"
fi
log "Version: $VERSION"

# Step 1: Tests
if [[ "$SKIP_TEST" == false ]]; then
    log "Step 1/4: Running tests..."
    bash scripts/test.sh --coverage
    ok "Tests passed"
else
    warn "Step 1/4: Tests skipped"
fi

# Step 2: Build
if [[ "$SKIP_BUILD" == false ]]; then
    log "Step 2/4: Building binaries..."
    bash scripts/build.sh --version "$VERSION"
    ok "Build complete"
else
    warn "Step 2/4: Build skipped"
fi

# Step 3: Git tag
log "Step 3/4: Git tag..."
if git tag -l | grep -q "^${VERSION}$"; then
    warn "Tag $VERSION already exists, skipping tag creation"
else
    if [[ "$DRY_RUN" == true ]]; then
        log "[DRY RUN] Would create tag: $VERSION"
    else
        git tag -a "$VERSION" -m "Release $VERSION"
        ok "Created tag: $VERSION"
    fi
fi

# Step 4: Push
log "Step 4/4: Push..."
if [[ "$DRY_RUN" == true ]]; then
    log "[DRY RUN] Would push branch and tag"
    log "[DRY RUN] Would trigger goreleaser"
else
    read -p "Push to remote and create release? [y/N] " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git push origin "$CURRENT_BRANCH"
        git push origin "$VERSION"
        ok "Pushed $VERSION to origin"

        # Trigger goreleaser if available
        if command -v goreleaser &>/dev/null; then
            log "Running goreleaser..."
            goreleaser release --clean
            ok "Release published"
        else
            warn "goreleaser not found, skipping release creation"
            log "Install: go install github.com/goreleaser/goreleaser@latest"
            log "Or push the tag to trigger GitHub Actions"
        fi
    else
        log "Push cancelled"
    fi
fi

echo ""
ok "Deploy pipeline complete"
