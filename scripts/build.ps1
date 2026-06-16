# bytemsg233 一键构建脚本 (PowerShell)
# Usage: pwsh scripts/build.ps1 [-Version v1.0.0] [-Os "linux,darwin,windows"] [-Arch "amd64,arm64"]

param(
    [string]$Version = "dev",
    [string]$Os = "linux,darwin,windows",
    [string]$Arch = "amd64,arm64",
    [string]$OutputDir = "dist"
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $ProjectRoot

function Log($msg) { Write-Host "[build] $msg" -ForegroundColor Cyan }
function Ok($msg)  { Write-Host "[BUILD OK] $msg" -ForegroundColor Green }

$commit = try { & git rev-parse --short HEAD 2>$null } catch { "none" }
$date = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$ldflags = "-s -w -X main.version=$Version -X main.commit=$commit -X main.date=$date"

Log "Version: $Version"
Log "Commit:  $commit"
Log "Date:    $date"

Remove-Item -Recurse -Force $OutputDir -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

$osList = $Os -split ','
$archList = $Arch -split ','
$fail = 0

foreach ($o in $osList) {
    foreach ($a in $archList) {
        if ($o -eq "darwin" -and $a -eq "386") { continue }
        if ($o -eq "windows" -and $a -eq "arm64") { continue }

        $ext = if ($o -eq "windows") { ".exe" } else { "" }
        $binName = "bytemsg233_${o}_${a}${ext}"

        Log "Building $binName..."
        $env:GOOS = $o
        $env:GOARCH = $a

        & go build -ldflags "$ldflags" -trimpath -o "$OutputDir/$binName" ./cmd/bytemsg233
        if ($LASTEXITCODE -eq 0) {
            $size = (Get-Item "$OutputDir/$binName").Length / 1MB
            Ok "$binName ($([math]::Round($size, 2)) MB)"
        } else {
            Write-Host "FAILED: $binName" -ForegroundColor Red
            $fail = 1
        }
    }
}

# Checksums
Log "Generating checksums..."
$files = Get-ChildItem "$OutputDir" -File | Where-Object { $_.Name -ne "checksums.txt" }
$checksums = @()
foreach ($f in $files) {
    $hash = (Get-FileHash $f.FullName -Algorithm SHA256).Hash.ToLower()
    $checksums += "$hash  $($f.Name)"
}
$checksums | Set-Content "$OutputDir/checksums.txt"
Ok "checksums.txt"

# Summary
Write-Host ""
Log "=== Build Summary ==="
Get-ChildItem "$OutputDir" -File | ForEach-Object {
    $sizeMB = [math]::Round($_.Length / 1MB, 2)
    Write-Host "  $sizeMB MB  $($_.Name)"
}

if ($fail -eq 0) { Ok "All builds succeeded" } else { Write-Host "Some builds failed" -ForegroundColor Red; exit 1 }
