# Performance

> Go benchmark snapshot · Windows amd64 · AMD Ryzen 9 7900X3D

bytemsg233 is optimized for schema-driven binary payloads with small field headers, varint integers, zigzag signed integers, and no repeated field names. The goal is not to beat Protobuf in every microcase. The goal is a compact wire format with simpler JSON schema authoring, native language APIs, object pooling, localized comments, and strong large-message behavior.

## Payload Size

Lower is better.

| Scenario | bytemsg233 | Protobuf | MessagePack | JSON |
|---|---:|---:|---:|---:|
| Player profile, 10 fields | **61 B** | 61 B | 155 B | 173 B |
| Chat message, 5 fields | **57 B** | 57 B | 103 B | 116 B |
| Battle input, 10 players x 8 fields | **247 B** | 266 B | 931 B | 1,097 B |
| Leaderboard, 100 rows x 6 fields | **3,409 B** | 3,608 B | 8,711 B | 9,602 B |

| Scenario | vs Protobuf | vs MessagePack | vs JSON |
|---|---:|---:|---:|
| Player profile | 0% | -60.6% | -64.7% |
| Chat message | 0% | -44.7% | -50.9% |
| Battle input | -7.1% | -73.5% | -77.5% |
| Leaderboard | -5.5% | -60.9% | -64.5% |

## Encode Speed

Lower ns/op is better.

| Scenario | bytemsg233 | Protobuf | MessagePack | JSON |
|---|---:|---:|---:|---:|
| Player profile | 140 | **90** | 513 | 387 |
| Chat message | 154 | **107** | 375 | 317 |
| Battle input | **979** | 2,030 | 3,994 | 2,836 |
| Leaderboard | **9,277** | 26,729 | 52,826 | 21,990 |

Interpretation:

- Protobuf is faster on tiny encode cases.
- bytemsg233 pulls ahead as payloads become larger and repeated structures dominate.
- MessagePack and JSON pay heavily for map-like payload shape and field names.

## Decode Speed

Lower ns/op is better.

| Scenario | bytemsg233 | Protobuf | MessagePack | JSON |
|---|---:|---:|---:|---:|
| Player profile | 279 | **104** | 612 | 1,636 |
| Chat message | 224 | **86** | 349 | 969 |
| Battle input | 1,001 | - | **90** | 172 |

Decode still has room for generated fast paths. Current numbers are useful as a baseline, not the ceiling.

## Allocations

Lower allocs/op is better.

### Encode

| Scenario | bytemsg233 | Protobuf | MessagePack | JSON |
|---|---:|---:|---:|---:|
| Player profile | 2 | 3 | 4 | 1 |
| Battle input | 2 | 36 | 7 | 2 |
| Leaderboard | 2 | 394 | 11 | 2 |

### Decode

| Scenario | bytemsg233 | Protobuf | MessagePack | JSON |
|---|---:|---:|---:|---:|
| Player profile | 5 | 2 | 1 | 4 |
| Chat message | 5 | 2 | 1 | 4 |

Generated object pools are separate from these raw codec benchmark numbers. They are designed to reduce application-level churn after code generation, especially in client loops and Unity-style gameplay code.

## Run Locally

```bash
# Payload size comparison
go test ./pkg/binary/... -run "TestBenchmark_SizeComparison" -v

# Encoding benchmarks
go test ./pkg/binary/... -bench="BenchmarkEncode_" -benchmem

# Decoding benchmarks
go test ./pkg/binary/... -bench="BenchmarkDecode_" -benchmem

# Full benchmark set
go test ./pkg/binary/... -bench="Benchmark(Encode|Decode)_" -benchmem
```

## JSON Schema Used By New Examples

```json
{
  "schema": "bymsg/v1",
  "package": "com.example.benchmark",
  "PlayerProfile": {
    "id": { "type": "uint64", "tag": 1 },
    "name": { "type": "string", "tag": 2 },
    "level": { "type": "uint32", "tag": 3 },
    "exp": { "type": "uint64", "tag": 4 },
    "tags": { "type": "list<string>", "tag": 5 },
    "attrs": { "type": "map<string, string>", "tag": 6 }
  }
}
```

## Summary

bytemsg233 is strongest when the project needs all of these at once:

- payload size close to Protobuf and far below JSON/MessagePack;
- generated APIs that feel native in Go, C#, Java, TypeScript, and Python;
- built-in object pooling for client-heavy workloads;
- JSON schema files that work in normal editors and GitHub without custom plugins;
- localized class and field comments from the schema itself.
