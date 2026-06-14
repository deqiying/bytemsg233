# bytemsg233 CI 全流程脚本 (PowerShell)
# Usage: pwsh scripts/ci.ps1 [-Version v1.0.0]

param(
    [string]$Version = "dev"
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $ProjectRoot

function Log($msg)  { Write-Host "[ci] $msg" -ForegroundColor Cyan }
function Ok($msg)   { Write-Host "[CI OK] $msg" -ForegroundColor Green }
function Fail($msg) { Write-Host "[CI FAIL] $msg" -ForegroundColor Red; exit 1 }

Log "=== bytemsg233 CI Pipeline ==="
Log "Version: $Version"
Write-Host ""

# 1. Lint
Log "Step 1/6: Lint..."
$golangci = Get-Command golangci-lint -ErrorAction SilentlyContinue
if ($golangci) {
    & golangci-lint run ./...
    if ($LASTEXITCODE -ne 0) { Fail "Lint failed" }
} else {
    Write-Host "  golangci-lint not found, running go vet"
    & go vet ./...
    if ($LASTEXITCODE -ne 0) { Fail "Vet failed" }
}
Ok "Lint passed"

# 2. Test
Log "Step 2/6: Test..."
& go test ./... -race -count=1
if ($LASTEXITCODE -ne 0) { Fail "Tests failed" }
Ok "Tests passed"

# 3. Coverage
Log "Step 3/6: Coverage..."
& go test ./... -coverprofile=coverage.out -covermode=atomic 2>$null
$coverLine = & go tool cover -func=coverage.out 2>$null | Select-Object -Last 1
Write-Host "  $coverLine"

# 4. Size comparison
Log "Step 4/6: Size comparison..."
& go test ./pkg/binary/... -run "TestSizeComparison" -v 2>&1 | Select-String "(bytes|节省|ByteMsg|Protobuf|MsgPack|JSON)"

# 5. Build
Log "Step 5/6: Cross-platform build..."
& pwsh scripts/build.ps1 -Version $Version
if ($LASTEXITCODE -ne 0) { Fail "Build failed" }

# 6. Verify CLI
Log "Step 6/6: CLI verification..."
$exe = "dist\bytemsg233_windows_amd64.exe"
if (Test-Path $exe) {
    & $exe version
    Ok "CLI works"
}

Write-Host ""
Ok "=== CI Pipeline Complete ==="
Write-Host ""
Log "Artifacts:"
Get-ChildItem dist -File | ForEach-Object {
    $sizeMB = [math]::Round($_.Length / 1MB, 2)
    Write-Host "  $sizeMB MB  $($_.Name)"
}
