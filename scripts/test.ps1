# bytemsg233 一键测试脚本 (PowerShell)
# Usage: pwsh scripts/test.ps1 [-Coverage] [-Bench] [-Race] [-Verbose]

param(
    [switch]$Coverage,
    [switch]$Bench,
    [switch]$Race,
    [switch]$Verbose,
    [string]$Packages = "./..."
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $ProjectRoot

function Log($msg)  { Write-Host "[test] $msg" -ForegroundColor Cyan }
function Ok($msg)   { Write-Host "[PASS] $msg" -ForegroundColor Green }
function Err($msg)  { Write-Host "[FAIL] $msg" -ForegroundColor Red }

# Clean previous coverage
Remove-Item -Force coverage.out, coverage.html -ErrorAction SilentlyContinue

# Build test flags
$testFlags = @()
if ($Verbose)  { $testFlags += "-v" }
if ($Race)     { $testFlags += "-race" }
if ($Coverage) { $testFlags += "-coverprofile=coverage.out", "-covermode=atomic" }

# Run tests
Log "Running tests..."
$testResult = & go test @testFlags $Packages 2>&1
if ($LASTEXITCODE -eq 0) {
    Ok "All tests passed"
} else {
    $testResult | Write-Host
    Err "Tests failed"
    exit 1
}

# Coverage report
if ($Coverage) {
    Log "Generating coverage report..."
    $coverFunc = & go tool cover -func=coverage.out 2>&1
    $coverFunc | Select-Object -Last 1 | Write-Host
    & go tool cover -html=coverage.out -o coverage.html 2>$null
    if (Test-Path coverage.html) { Log "HTML report: coverage.html" }
}

# Benchmarks
if ($Bench) {
    Log "Running benchmarks..."
    & go test ./pkg/binary/... -bench=. -benchmem -count=3
}

# Size comparison
Log "Running size comparison..."
& go test ./pkg/binary/... -run "TestSizeComparison" -v

Log "Done."
