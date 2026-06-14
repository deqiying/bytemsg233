# bytemsg233 体积与性能对比

> 测试环境: Go 1.26, Windows amd64, AMD Ryzen 9 7900X3D

## 一、通用数据结构

### 单条记录 (UserProfile: id + name + email + tags + metadata)

| 格式 | 字节数 | vs ByteMsg |
|------|--------|------------|
| **ByteMsg** | **96 B** | 100% |
| Protobuf v3 | 101 B | 105.0% |
| MessagePack | 121 B | 126.0% |
| JSON | 152 B | 158.3% |

### 整数密集型 (4 fields × 4 records)

| 格式 | 字节数 | vs ByteMsg |
|------|--------|------------|
| **ByteMsg** | **52 B** | 100% |
| Protobuf v3 | 52 B | 100% |
| MessagePack | 133 B | 255.8% |
| JSON | 153 B | 294.2% |

---

## 二、游戏业务场景

### 场景 1: 登录推送 (Login Push)

> 玩家登录时服务端一次性下发的全量数据
> 含 30 英雄、80 背包物品、15 邮件、20 任务

| 格式 | 字节数 | vs ByteMsg |
|------|--------|------------|
| **ByteMsg** | **见基准** | — |
| MessagePack | 见基准 | — |
| JSON | 见基准 | — |

```bash
go test ./pkg/binary/... -run "TestGame_LoginPush" -v
```

### 场景 2: 战斗帧同步 (Battle Frame Sync)

> 实时对战每帧广播，10 玩家输入，30fps

| 格式 | 字节数 | vs ByteMsg |
|------|--------|------------|
| **ByteMsg** | **284 B** | 100% |
| MessagePack | 846 B | 297.9% |
| JSON | 961 B | 338.4% |

**ByteMsg 比 JSON 小 70.4%，30fps 带宽节省显著**

### 场景 3: 排行榜 (Leaderboard — 100 players)

```bash
go test ./pkg/binary/... -run "TestGame_Leaderboard" -v
```

### 场景 4: 聊天消息 (Chat Message)

| 格式 | 字节数 | vs ByteMsg |
|------|--------|------------|
| **ByteMsg** | **61 B** | 100% |
| MessagePack | 106 B | 173.8% |
| JSON | 119 B | 195.1% |

### 场景 5: 背包 200 件物品

| 格式 | 字节数 | vs ByteMsg |
|------|--------|------------|
| **ByteMsg** | **1,733 B** | 100% |
| MessagePack | 8,803 B | 507.9% |
| JSON | 10,864 B | 626.8% |

**ByteMsg 比 JSON 小 84.0%**

---

## 三、游戏场景总结

| 场景 | ByteMsg | MsgPack | JSON | vs JSON | vs MsgPack |
|------|---------|---------|------|---------|------------|
| 英雄数据 | 183 B | 491 B | 561 B | **-67.4%** | -62.7% |
| 聊天消息 | 61 B | 106 B | 119 B | **-48.7%** | -42.5% |
| 战斗帧 | 284 B | 846 B | 961 B | **-70.4%** | -66.4% |
| 背包 200 件 | 1,733 B | 8,803 B | 10,864 B | **-84.0%** | -80.3% |
| 批量聊天 100 条 | 9,100 B | 13,603 B | 15,001 B | **-39.3%** | -33.1% |

---

## 四、编码性能

> 测试环境: AMD Ryzen 9 7900X3D, Go 1.26, Windows amd64

### 编码速度 (Encode, ns/op — 越小越好)

| 场景 | ByteMsg | MessagePack | JSON | ByteMsg vs JSON |
|------|---------|-------------|------|-----------------|
| 登录推送 (30英雄+80背包) | **24,368** | 68,780 | 47,622 | **1.96x 快** |
| 战斗帧 (10人, 高频同步) | **1,723** | 4,337 | 3,188 | **1.85x 快** |
| 排行榜 (100人) | **30,606** | 73,018 | 55,127 | **1.80x 快** |

### 内存分配 (allocs/op — 越小越好)

| 场景 | ByteMsg | MessagePack | JSON |
|------|---------|-------------|------|
| 登录推送 | 219 | 191 | 212 |
| 战斗帧 | 9 | 7 | 2 |
| 排行榜 | 13 | 11 | 2 |

### 体积 + 速度综合对比

| 场景 | 格式 | 字节数 | 编码 ns/op | 字节×ns |
|------|------|--------|-----------|---------|
| 登录推送 | **ByteMsg** | **小** | **快 1.96x** | 最优 |
| 登录推送 | JSON | 大 | 慢 | 最差 |
| 战斗帧 | **ByteMsg** | **小 70%** | **快 1.85x** | 最优 |
| 战斗帧 | JSON | 大 | 慢 | 最差 |
| 排行榜 | **ByteMsg** | **小** | **快 1.80x** | 最优 |
| 排行榜 | JSON | 大 | 慢 | 性能体积都输 |

### 优化说明

ByteMsg 编码器使用两个关键优化：

1. **栈分配 varint buffer** — `Encoder.buf [10]byte` 替代 `make([]byte, 10)`，消除每次 WriteVarint 的堆分配
2. **Buffer 池化** — `sync.Pool` 复用嵌套结构的临时 buffer

```bash
go test ./pkg/binary/... -bench="BenchmarkGame" -benchmem
```
