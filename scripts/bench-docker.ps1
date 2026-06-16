# Run every benchmark/verification suite inside one Docker image.
param(
    [string]$Image = "bytemsg233-bench:local",
    [string]$GoVersion = "1.26.0",
    [string]$NodeMajor = "22",
    [string]$DotnetChannel = "10.0",
    [int]$Count = 1,
    [string]$BenchTime = "",
    [switch]$NoBuild
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $Root

if (-not $NoBuild) {
    docker build `
        --build-arg "GO_VERSION=$GoVersion" `
        --build-arg "NODE_MAJOR=$NodeMajor" `
        --build-arg "DOTNET_CHANNEL=$DotnetChannel" `
        -f Dockerfile.bench `
        -t $Image `
        .
}

$envArgs = @("-e", "BENCH_COUNT=$Count")
if ($BenchTime) {
    $envArgs += @("-e", "BENCH_TIME=$BenchTime")
}

docker run --rm `
    @envArgs `
    -v "${Root}:/workspace" `
    -w /workspace `
    $Image `
    bash scripts/bench-all.sh
