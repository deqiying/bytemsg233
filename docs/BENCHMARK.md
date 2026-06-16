# Performance

> Go benchmark snapshot · Windows amd64 · AMD Ryzen 9 9950X

This page explains what the numbers mean, not just who wins a table.

ByteMsg233 is optimized for schema-driven game and client traffic: small field headers, varint integers, zigzag signed integers, no repeated field names, and generated APIs that can reuse memory. The target is a practical protocol: readable schema, compact packets, native generated code, high-throughput encode/decode on repeated DTO payloads, and a zero-GC hot path where the caller provides buffers and pools are prewarmed.

The headline is intentionally modest: low allocation pressure, stable frame-time behavior, compact bytes, and strong repeated-structure throughput. Tiny packets can still favor mature specialized implementations; larger client payloads are where the design should show up most clearly.

## How To Read The Tables

Lower is better for size, `ns/op`, `B/op`, and `allocs/op`.

Comparison order is fixed:

1. ByteMsg233
2. Protobuf
3. JSON
4. Optional extra codecs, such as MessagePack

JSON is included because many teams start there. MessagePack is included because it is a common "binary JSON" baseline. Protobuf is included because it is the obvious mature competitor.

## Payload Size

Payload size matters most when the same shape repeats: rankings, inventory rows, battle inputs, quest lists, mail lists, and state snapshots.

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile, 10 fields | **61 B** | 61 B | 173 B | 155 B |
| Chat message, 5 fields | **57 B** | 57 B | 116 B | 103 B |
| ChatDto all types, list/map/custom | **304 B** | 316 B | 647 B | 531 B |
| Battle input, 10 players x 8 fields | **247 B** | 266 B | 1097 B | 931 B |
| TaskDto list, 100 rows x 9 fields | **3845 B** | 4044 B | 14691 B | 13303 B |
| Leaderboard, 100 rows x 6 fields | **3409 B** | 3608 B | 9602 B | 8711 B |

Savings versus other codecs:

| Scenario | vs Protobuf | vs JSON | vs MessagePack |
|---|---:|---:|---:|
| Player profile | 0% | -64.7% | -60.6% |
| Chat message | 0% | -50.9% | -44.7% |
| ChatDto all types | -3.8% | -53.0% | -42.7% |
| Battle input | -7.1% | -77.5% | -73.5% |
| TaskDto list | -4.9% | -73.8% | -71.1% |
| Leaderboard | -5.5% | -64.5% | -60.9% |

## Encode Speed

Tiny packets are where mature libraries like Protobuf can still win. Bigger repeated structures are where ByteMsg233 becomes more interesting.

These values are duration. Lower `ns/op` is better.

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **137.8** | 139.1 | 390.0 | 477.3 |
| Chat message | 143.4 | **89.6** | 281.9 | 337.4 |
| ChatDto all types | **249.8** | 1005 | 2175 | 2266 |
| Battle input | **822.6** | 1681 | 2228 | 3513 |
| TaskDto list, 100 rows | **3231** | 23374 | 27944 | 46550 |
| Leaderboard | **7367** | 23444 | 17085 | 39710 |

The same ChatDto result as throughput. Higher `ops/s` is better.

| Codec | Encode ops/s | Decode ops/s |
|---|---:|---:|
| ByteMsg233 | **4003203** | 493097 |
| Protobuf | 995025 | **914077** |
| JSON | 459770 | 151630 |
| MessagePack | 441306 | 444050 |

ChatDto relative view:

| Codec | Encode duration | Decode duration | Encode throughput | Decode throughput |
|---|---:|---:|---:|---:|
| ByteMsg233 | **0.25x Protobuf** | 1.85x Protobuf | **4.02x Protobuf** | 0.54x Protobuf |
| Protobuf | 4.02x ByteMsg233 | **0.54x ByteMsg233** | 0.25x ByteMsg233 | **1.85x ByteMsg233** |
| JSON | 8.71x ByteMsg233 | 6.03x Protobuf | 0.11x ByteMsg233 | 0.17x Protobuf |
| MessagePack | 9.07x ByteMsg233 | 2.06x Protobuf | 0.11x ByteMsg233 | 0.49x Protobuf |

Interpretation:

- Protobuf is still excellent on tiny decode cases.
- ByteMsg233 encode uses the append hot path: caller-owned buffer, precomputed nested sizes, no temporary nested buffers, `0 B/op`.
- ByteMsg233 pulls ahead when repeated structures dominate.
- JSON and MessagePack pay for dynamic object shape and field-name-heavy data.
- The performance goal for generated decode is the same shape: reusable state, caller-owned storage where practical, and no hot-path GC after pool prewarm.

## Decode Speed

Decode numbers are a baseline. Generated fast paths and pool-aware decoders are expected to improve this area.

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | 226.9 | **88.7** | 1345 | 552.5 |
| Chat message | 198.0 | **79.1** | 764.8 | 279.4 |
| ChatDto all types | 2028 | **1094** | 6595 | 2252 |
| Battle input | 889.1 | - | 156.4 | **70.6** |

## Allocations

Allocations are where game clients feel pain: a small per-packet allocation can become a frame-time spike when repeated thousands of times.

### Encode (`B/op`, `allocs/op`)

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **64, 1** | 104, 3 | 176, 1 | 496, 4 |
| ChatDto all types | **0, 0** | 1328, 22 | 1282, 11 | 2323, 7 |
| Battle input | **288, 2** | 1560, 36 | 1177, 2 | 2059, 7 |
| TaskDto list, 100 rows | **0, 0** | 23160, 410 | 16446, 2 | 32830, 11 |
| Leaderboard | **3488, 2** | 22136, 394 | 9773, 2 | 32830, 11 |

### Decode (`B/op`, `allocs/op`)

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | 128, 5 | **40, 2** | 216, 4 | 48, 1 |
| Chat message | 160, 5 | **56, 2** | 216, 4 | 48, 1 |
| ChatDto all types | 1536, 56 | 752, 26 | 600, 28 | **296, 18** |
| Battle input | **48, 1** | - | 144, 1 | **48, 1** |

Generated object pools are separate from these raw codec benchmark numbers. They reduce application-level churn after code generation, especially in Unity-style gameplay code and client update loops. Runtime pools are single-threaded and lock-free by policy so hot-path memory reuse stays predictable.

For hot-path encode code, prefer caller-owned buffers. `AppendEncoder` is the zero-GC path for preallocated byte slices:

```bash
go test ./pkg/binary -run ^$ -bench "BenchmarkEncode_TaskList" -benchtime=1000x -benchmem
```

Current `TaskList_ByteMsg233` hot-path target: `0 B/op`, `0 allocs/op` for 100 `TaskDto` entries.

## Game Traffic Coverage

The benchmark suite must cover real packet families, not only a business DTO list.

| Scenario | Structure |
|---|---|
| Login push | player, 30 heroes, 80 items, 15 mails, 20 quests, settings |
| Battle frame | 10 player inputs, frame id, timestamp, random seed |
| ChatDto all types | bool, signed/unsigned ints, float, double, string, bytes, list, map KV, nested custom messages |
| Leaderboard | 100 rank rows with player, guild, avatar, score |
| Battle input | compact input batch with fixed numeric fields |
| TaskDto list | 100 business DTO rows for non-game repeated data |

Run the game packet checks:

```bash
go test ./pkg/binary -run "TestGame_" -v
go test ./pkg/binary -run "TestBenchmark_ChatDtoAllTypesRoundTrip" -v
go test ./pkg/binary -run ^$ -bench "BenchmarkGame_" -benchmem
go test ./pkg/binary -run ^$ -bench "Benchmark(Encode|Decode)_ChatDtoAllTypes" -benchmem
```

See [GAME_BINARY.md](GAME_BINARY.md) for the message-shape rules.

## Run Locally

```bash
# Payload size comparison
go test ./pkg/binary/... -run "TestBenchmark_SizeComparison" -v

# Game packet checks
go test ./pkg/binary/... -run "TestGame_" -v

# Encoding benchmarks
go test ./pkg/binary/... -bench="BenchmarkEncode_" -benchmem

# Decoding benchmarks
go test ./pkg/binary/... -bench="BenchmarkDecode_" -benchmem

# Game benchmarks
go test ./pkg/binary/... -bench="BenchmarkGame_" -benchmem
```

## Summary

ByteMsg233 is strongest when the project needs all of these at once:

- packet size close to Protobuf and far below JSON/MessagePack;
- generated APIs that feel native in Go, C#, Java, TypeScript, Rust, C++, C, Kotlin, Swift, Dart, Lua, and Python;
- object pooling for client-heavy workloads;
- JSON schema files that are readable in normal editors;
- debug-friendly text output outside the hot path.
