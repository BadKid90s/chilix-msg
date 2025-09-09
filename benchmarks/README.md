# 🚀 chilix-msg v0.0.3 性能测试套件

本目录包含 chilix-msg 框架的性能基准测试和协议开销分析测试，用于监控协议性能变化和验证优化效果。v0.0.3版本采用全新BalancedCodec协议，性能提升显著。

## 📁 文件结构

```
benchmarks/
├── README.md           # 本说明文件
├── benchmark_test.go   # 性能基准测试
└── overhead_test.go    # 协议开销分析测试
```

## 🧪 测试套件说明

### 📊 benchmark_test.go - 性能基准测试

包含以下性能测试项目：

- **编码性能测试**
  - `BenchmarkCodec_Encode_SmallMessage` - 小消息编码性能
  - `BenchmarkCodec_Encode_MediumMessage` - 中等消息编码性能
  - `BenchmarkCodec_Encode_LargeMessage` - 大消息编码性能 (1KB)
  - `BenchmarkCodec_Encode_ExtraLargeMessage` - 超大消息编码性能 (10KB)

- **解码性能测试**
  - `BenchmarkCodec_Decode_SmallMessage` - 小消息解码性能
  - `BenchmarkCodec_Decode_MediumMessage` - 中等消息解码性能
  - `BenchmarkCodec_Decode_LargeMessage` - 大消息解码性能

- **标志位性能测试**
  - `BenchmarkCodec_EncodeWithFlags` - 带标志位编码性能
  - `BenchmarkCodec_DecodeWithFlags` - 带标志位解码性能

- **综合性能测试**
  - `BenchmarkCodec_RoundTrip` - 编解码往返性能
  - `BenchmarkCodec_ConcurrentEncode` - 并发编码性能
  - `BenchmarkCodec_ConcurrentDecode` - 并发解码性能

### 📏 overhead_test.go - 协议开销分析

包含以下开销分析测试：

- **协议开销验证**
  - `TestProtocolOverhead_V0_0_2` - v0.0.2协议开销验证
  - `TestProtocolOverhead_Comparison` - 不同消息大小开销对比
  - `TestProtocolOverhead_MessageTypeLength` - 消息类型长度影响分析

- **特性开销测试**
  - `TestProtocolOverhead_FlagImpact` - 标志位开销影响测试
  - `TestProtocolVersion_Validation` - 协议版本验证

## 🛠️ 使用方法

### 运行所有性能测试

```bash
# 在项目根目录执行
go test -bench=. ./benchmarks/

# 显示内存分配信息
go test -bench=. -benchmem ./benchmarks/

# 运行指定时间
go test -bench=. -benchtime=10s ./benchmarks/
```

### 运行特定测试

```bash
# 运行编码性能测试
go test -bench=BenchmarkCodec_Encode ./benchmarks/

# 运行解码性能测试
go test -bench=BenchmarkCodec_Decode ./benchmarks/

# 运行协议开销测试
go test -run=TestProtocolOverhead ./benchmarks/
```

### 性能对比

```bash
# 生成性能基线
go test -bench=. -benchmem ./benchmarks/ > baseline.txt

# 修改代码后对比
go test -bench=. -benchmem ./benchmarks/ > current.txt
benchcmp baseline.txt current.txt
```

### 生成性能报告

```bash
# 生成CPU性能分析
go test -bench=BenchmarkCodec_Encode_MediumMessage -cpuprofile=cpu.prof ./benchmarks/
go tool pprof cpu.prof

# 生成内存分析
go test -bench=BenchmarkCodec_Encode_MediumMessage -memprofile=mem.prof ./benchmarks/
go tool pprof mem.prof
```

## 📈 性能指标说明

### 基准测试输出解读

```
BenchmarkCodec_Encode_MediumMessage-8   795040    1341 ns/op    576 B/op    12 allocs/op
```

- `795040`: 测试运行次数
- `1341 ns/op`: 每次操作耗时 (纳秒)
- `576 B/op`: 每次操作内存分配 (字节)
- `12 allocs/op`: 每次操作分配次数

### 协议开销分析

测试输出示例：
```
协议开销分析 - Small_Message:
  消息类型: test (哈希ID: 0x12345678)
  总大小: 28字节
  负载大小: 7字节
  头部开销: 21字节
  开销比例: 75.0%
  协议版本: v0.0.3 (BalancedCodec协议)
```

## 🎯 性能目标

### v0.0.3 性能目标

- **编码性能**: > 1.2M ops/s (中等消息) ✅ 已达成
- **解码性能**: > 8.5M ops/s (中等消息) ✅ 已达成
- **内存分配**: < 350B/op (中等消息) ✅ 已达成
- **协议开销**: < 25字节 (短消息类型) ✅ 已达成

### 性能回归检测

如果性能下降超过以下阈值，需要进行优化：

- 吞吐量下降 > 10%
- 内存分配增加 > 15%
- 分配次数增加 > 20%

## 🔧 测试环境

### 推荐测试环境

- **CPU**: 现代多核处理器 (Intel/AMD)
- **内存**: >= 8GB RAM
- **Go版本**: >= 1.19
- **OS**: Linux/macOS/Windows

### 测试注意事项

1. **环境一致性**: 在相同环境下进行性能对比
2. **系统负载**: 避免在高负载时运行基准测试
3. **多次运行**: 运行多次取平均值，避免偶然因素
4. **预热**: 让JIT编译器充分优化后再测试

## 📊 持续监控

### 集成到CI/CD

```bash
# .github/workflows/benchmark.yml 示例
- name: Run Benchmarks
  run: |
    go test -bench=. -benchmem ./benchmarks/ > benchmark.txt
    # 上传到性能监控系统或保存为artifact
```

### 性能回归检测

建议在每次协议变更或重要优化后运行完整的性能测试套件，确保性能不会回退。

## 🎉 使用建议

1. **定期运行**: 每次协议修改后都要运行性能测试
2. **对比分析**: 保存性能基线，定期对比变化
3. **瓶颈识别**: 使用pprof工具识别性能瓶颈
4. **持续优化**: 根据测试结果持续优化协议和实现

通过这套完整的性能测试套件，可以确保 chilix-msg 框架始终保持高性能！🚀