#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

STAMP="${BENCH_STAMP:-$(date -u +%Y%m%dT%H%M%SZ)}"
OUT="${BENCH_OUT:-bench-results/${STAMP}}"
COUNT="${BENCH_COUNT:-1}"
TIME="${BENCH_TIME:-}"
mkdir -p "$OUT"

log() {
  printf '\n[bench] %s\n' "$*"
}

run() {
  local name="$1"
  shift
  log "$name"
  "$@" 2>&1 | tee "${OUT}/${name}.log"
}

run_sh() {
  local name="$1"
  shift
  log "$name"
  bash -lc "$*" 2>&1 | tee "${OUT}/${name}.log"
}

GO_BENCH_FLAGS=(-run '^$' -benchmem -count="$COUNT")
if [[ -n "$TIME" ]]; then
  GO_BENCH_FLAGS+=(-benchtime="$TIME")
fi

{
  echo "timestamp=${STAMP}"
  echo "os=$(uname -a)"
  echo "go=$(go version)"
  echo "node=$(node --version)"
  echo "npm=$(npm --version)"
  echo "rust=$(rustc --version)"
  echo "cargo=$(cargo --version)"
  echo "dotnet=$(dotnet --version)"
  echo "java=$(java -version 2>&1 | head -n 1)"
  echo "javac=$(javac -version 2>&1)"
} | tee "${OUT}/versions.txt"

run go-tests go test ./... -count=1
run go-size go test ./pkg/binary -run 'TestBenchmark_SizeComparison|TestBenchmark_ChatDtoAllTypesRoundTrip|TestGame_' -v
run go-encode-bench go test ./pkg/binary "${GO_BENCH_FLAGS[@]}" -bench 'BenchmarkEncode_'
run go-decode-bench go test ./pkg/binary "${GO_BENCH_FLAGS[@]}" -bench 'BenchmarkDecode_'
run go-game-bench go test ./pkg/binary "${GO_BENCH_FLAGS[@]}" -bench 'BenchmarkGame_'
run go-hotpath-bench go test ./pkg/binary "${GO_BENCH_FLAGS[@]}" -bench 'BenchmarkEncode_(ChatDtoAllTypes|TaskList)_ByteMsg233'

run_sh typescript-runtime "cd libs/typescript && npm test"
run_sh rust-runtime "cd libs/rust && cargo test"
run_sh csharp-runtime "cd libs/csharp && dotnet run --project ./Tests/ByteMsg233.Tests.csproj"

run_sh java-runtime-javac "cd libs/java && rm -rf build/javac && mkdir -p build/javac && javac --release 17 -d build/javac \$(find src/main/java -name '*.java' | sort) && rm -rf build/javac"
run_sh java-generated-javac "rm -rf .bench-java-gen .bench-java-classes && go run ./cmd/bytemsg233 compile ./testdata/user.json --lang java -o .bench-java-gen && mkdir -p .bench-java-classes && javac --release 17 -d .bench-java-classes \$(find libs/java/src/main/java .bench-java-gen -name '*.java' | sort) && rm -rf .bench-java-gen .bench-java-classes"

log "done: ${OUT}"
