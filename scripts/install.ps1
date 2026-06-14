#Requires -Version 5.0
param(
    [string]$Version = "latest"
)

$ErrorActionPreference = "Stop"

$Repo = "neko233-com/bytemsg233"
$Binary = "bytemsg233"
$InstallDir = "$env:LOCALAPPDATA\bytemsg233"

if ($Version -eq "latest") {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $release.tag_name
}

$url = "https://github.com/$Repo/releases/download/$Version/${Binary}_windows_amd64.zip"

Write-Host "Downloading $Binary $Version for Windows..."
$tmpFile = "$env:TEMP\bytemsg233.zip"
Invoke-WebRequest -Uri $url -OutFile $tmpFile

Write-Host "Installing to $InstallDir..."
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

Expand-Archive -Path $tmpFile -DestinationPath $InstallDir -Force
Remove-Item $tmpFile

$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
}

Write-Host "$Binary $Version installed successfully!"
Write-Host "Please restart your terminal to use $Binary."
