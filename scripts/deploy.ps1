# bytemsg233 发布脚本 (PowerShell)
# Usage: pwsh scripts/deploy.ps1 [-Version v1.0.0] [-DryRun] [-SkipTest]

param(
    [string]$Version = "",
    [switch]$DryRun,
    [switch]$SkipTest,
    [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $ProjectRoot

function Log($msg)  { Write-Host "[deploy] $msg" -ForegroundColor Cyan }
function Ok($msg)   { Write-Host "[DEPLOY OK] $msg" -ForegroundColor Green }
function Warn($msg) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }

# Validate git state
$status = & git status --porcelain 2>$null
if ($status) {
    Warn "Working tree has uncommitted changes"
    & git status --short
}

$branch = & git branch --show-current
Log "Current branch: $branch"

if (!$Version) {
    $Version = try { & git describe --tags --always 2>$null } catch { "dev" }
}
Log "Version: $Version"

# Step 1: Tests
if (!$SkipTest) {
    Log "Step 1/4: Running tests..."
    & pwsh scripts/test.ps1 -Coverage
    Ok "Tests passed"
} else {
    Warn "Step 1/4: Tests skipped"
}

# Step 2: Build
if (!$SkipBuild) {
    Log "Step 2/4: Building binaries..."
    & pwsh scripts/build.ps1 -Version $Version
    Ok "Build complete"
} else {
    Warn "Step 2/4: Build skipped"
}

# Step 3: Git tag
Log "Step 3/4: Git tag..."
$tags = & git tag -l 2>$null
if ($tags -contains $Version) {
    Warn "Tag $Version already exists, skipping"
} else {
    if ($DryRun) {
        Log "[DRY RUN] Would create tag: $Version"
    } else {
        & git tag -a $Version -m "Release $Version"
        Ok "Created tag: $Version"
    }
}

# Step 4: Push
Log "Step 4/4: Push..."
if ($DryRun) {
    Log "[DRY RUN] Would push branch and tag"
} else {
    $confirm = Read-Host "Push to remote and create release? [y/N]"
    if ($confirm -eq 'y' -or $confirm -eq 'Y') {
        & git push origin $branch
        & git push origin $Version
        Ok "Pushed $Version to origin"

        $goreleaser = Get-Command goreleaser -ErrorAction SilentlyContinue
        if ($goreleaser) {
            Log "Running goreleaser..."
            & goreleaser release --clean
            Ok "Release published"
        } else {
            Warn "goreleaser not found, push tag to trigger GitHub Actions"
        }
    } else {
        Log "Push cancelled"
    }
}

Ok "Deploy pipeline complete"
