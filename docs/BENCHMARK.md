# Performance

> Go benchmark snapshot · Windows amd64 · AMD Ryzen 9 7900X3D

This page explains what the numbers mean, not just who wins a table.

ByteMsg233 is optimized for schema-driven game and client traffic: small field headers, varint integers, zigzag signed integers, no repeated field names, and generated APIs that can reuse memory. The target is a practical protocol: readable schema, compact packets, native generated code, high-throughput encode/decode on repeated DTO payloads, and a zero-GC hot path where the caller provides buffers and pools are prewarmed.

The headline is straightforward: the Go fast path now beats the Protobuf wire helper baselines in every Protobuf comparison in this suite, while keeping allocation pressure low and repeated client payloads compact.

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

Tiny packets and repeated structures both use the ByteMsg233 fast path in this snapshot.

These values are duration. Lower `ns/op` is better.

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **38.2** | 92.6 | 269.0 | 375.0 |
| Chat message | **30.1** | 71.0 | 191.9 | 229.8 |
| ChatDto all types | **254.8** | 1322 | 1695 | 1497 |
| Battle input | **181.7** | 1286 | 2007 | 2674 |
| TaskDto list, 100 rows | **3319** | 14699 | 24965 | 28419 |
| Leaderboard | **1938** | 14193 | 15822 | 21406 |

The same ChatDto result as throughput. Higher `ops/s` is better.

| Codec | Encode ops/s | Decode ops/s |
|---|---:|---:|
| ByteMsg233 | **3924647** | **1676165** |
| Protobuf | 756430 | 801282 |
| JSON | 589971 | 127340 |
| MessagePack | 668003 | 385802 |

ChatDto relative view:

| Codec | Encode duration | Decode duration | Encode throughput | Decode throughput |
|---|---:|---:|---:|---:|
| ByteMsg233 | **0.19x Protobuf** | **0.48x Protobuf** | **5.19x Protobuf** | **2.09x Protobuf** |
| Protobuf | 5.19x ByteMsg233 | 2.09x ByteMsg233 | 0.19x ByteMsg233 | 0.48x ByteMsg233 |
| JSON | 6.65x ByteMsg233 | 6.29x Protobuf | 0.15x ByteMsg233 | 0.16x Protobuf |
| MessagePack | 5.88x ByteMsg233 | 2.08x Protobuf | 0.17x ByteMsg233 | 0.48x Protobuf |

Interpretation:

- ByteMsg233 simple DTO encode uses append helpers; complex generated-style DTO encode uses caller-owned buffers and precomputed nested sizes.
- ByteMsg233 decode uses `SliceDecoder` and zero-copy string/bytes views for immutable payload buffers.
- In this suite, ByteMsg233 is faster than Protobuf for every measured Protobuf encode/decode comparison.
- JSON and MessagePack pay for dynamic object shape and field-name-heavy data.
- The performance goal for generated decode is reusable state, caller-owned storage where practical, and low hot-path GC after pool prewarm.

## Decode Speed

Decode uses the slice fast path. Lower `ns/op` is better.

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **52.0** | 111.8 | 1609 | 576.8 |
| Chat message | **29.4** | 93.0 | 914.9 | 351.9 |
| ChatDto all types | **596.6** | 1248 | 7853 | 2592 |
| Battle input | 382.7 | - | 161.3 | **93.4** |

## Allocations

Allocations are where game clients feel pain: a small per-packet allocation can become a frame-time spike when repeated thousands of times.

### Encode (`B/op`, `allocs/op`)

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **64, 1** | 104, 3 | 176, 1 | 496, 4 |
| ChatDto all types | **0, 0** | 1328, 22 | 1282, 11 | 2323, 7 |
| Battle input | **256, 1** | 1560, 36 | 1177, 2 | 2058, 7 |
| TaskDto list, 100 rows | **0, 0** | 23160, 410 | 16446, 2 | 32830, 11 |
| Leaderboard | **4096, 1** | 22136, 394 | 9766, 2 | 32809, 11 |

### Decode (`B/op`, `allocs/op`)

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **0, 0** | 40, 2 | 216, 4 | 48, 1 |
| Chat message | **0, 0** | 56, 2 | 216, 4 | 48, 1 |
| ChatDto all types | **432, 5** | 752, 26 | 600, 28 | 296, 18 |
| Battle input | **0, 0** | - | 144, 1 | 48, 1 |

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
