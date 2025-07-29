# MarketStore Rust客户端TDD开发计划

## 项目概述

基于MarketStore的gRPC + WebSocket混合架构，采用TDD（测试驱动开发）模式开发的高性能Rust客户端。

## 技术架构

### 核心组件
- **gRPC客户端**: 处理CRUD操作（查询、写入、创建、删除）
- **WebSocket客户端**: 处理实时数据流订阅
- **混合客户端**: 统一接口，结合两种协议的优势
- **数据模型**: 强类型的数据结构定义
- **错误处理**: 完善的错误类型和处理机制

### 技术栈
- **语言**: Rust 2021 Edition
- **异步运行时**: Tokio
- **gRPC**: Tonic + Prost
- **WebSocket**: Tokio-tungstenite
- **序列化**: Serde + MessagePack
- **测试**: 内置测试框架 + Mockall

## TDD开发周期

### 第一阶段：项目初始化与基础架构 ✅

#### 1.1 项目结构设置
```bash
marketstore-rust-client/
├── Cargo.toml                    # 依赖配置
├── build.rs                      # protobuf编译脚本
├── proto/
│   └── marketstore.proto         # 从MarketStore复制
├── src/
│   ├── lib.rs                    # 库入口
│   ├── main.rs                   # 示例程序
│   ├── client/                   # 客户端模块
│   ├── models/                   # 数据模型
│   ├── error/                    # 错误处理
│   └── utils/                    # 工具函数
└── tests/
    ├── unit/                     # 单元测试
    └── integration/              # 集成测试
```

#### 1.2 依赖配置
- **gRPC**: `tonic`, `prost`
- **WebSocket**: `tokio-tungstenite`, `futures`
- **异步**: `tokio`
- **序列化**: `serde`, `rmp-serde`
- **错误处理**: `anyhow`, `thiserror`
- **测试**: `tokio-test`, `mockall`

### 第二阶段：错误处理模块 ✅

#### 2.1 测试用例
- ✅ gRPC错误创建和转换
- ✅ WebSocket错误处理
- ✅ 传输错误处理
- ✅ 序列化错误处理
- ✅ 数据验证错误
- ✅ 连接错误
- ✅ 错误显示和转换

#### 2.2 实现功能
- ✅ `MarketStoreError` 枚举定义
- ✅ 错误类型转换实现
- ✅ 错误消息格式化
- ✅ 与第三方库错误集成

### 第三阶段：数据模型模块 ✅

#### 3.1 测试用例
- ✅ OHLCV数据结构创建和序列化
- ✅ 查询请求构建器
- ✅ 流订阅管理
- ✅ 数据形状定义
- ✅ 写入请求创建

#### 3.2 实现功能
- ✅ `OHLCVData` 结构体
- ✅ `QueryRequest` 和构建器模式
- ✅ `StreamSubscription` 流订阅
- ✅ `DataShape` 数据形状
- ✅ 序列化支持

### 第四阶段：gRPC客户端模块 ✅

#### 4.1 测试用例
- ✅ gRPC连接测试
- ✅ 数据转换测试
- ✅ 查询操作测试
- ✅ 写入操作测试
- ✅ 批量操作测试

#### 4.2 实现功能
- ✅ `GrpcClient` 结构体
- ✅ `GrpcClientTrait` 接口定义
- ✅ 数据转换函数
- ✅ protobuf集成
- ✅ 异步操作支持

### 第五阶段：WebSocket客户端模块 ✅

#### 5.1 测试用例
- ✅ WebSocket连接测试
- ✅ 消息序列化测试
- ✅ 流订阅测试
- ✅ 错误处理测试
- ✅ 取消机制测试

#### 5.2 实现功能
- ✅ `WebSocketClient` 结构体
- ✅ 实时数据订阅
- ✅ MessagePack序列化
- ✅ 流处理机制
- ✅ 取消支持

### 第六阶段：混合客户端模块 ✅

#### 6.1 测试用例
- ✅ 混合客户端创建
- ✅ 统一接口测试
- ✅ 批量操作测试
- ✅ 实时订阅测试
- ✅ 健康检查测试

#### 6.2 实现功能
- ✅ `MarketStoreClient` 统一接口
- ✅ gRPC和WebSocket集成
- ✅ 批量操作支持
- ✅ 健康检查机制
- ✅ 并发安全设计

## 测试策略

### 单元测试
- **错误处理**: 测试所有错误类型的创建和转换
- **数据模型**: 测试数据结构的创建、验证和序列化
- **gRPC客户端**: 测试连接、查询、写入等操作
- **WebSocket客户端**: 测试连接、订阅、消息处理

### 集成测试
- **混合客户端**: 测试gRPC和WebSocket的集成
- **批量操作**: 测试批量查询和写入
- **实时数据流**: 测试实时数据订阅和处理
- **错误恢复**: 测试连接断开和重连

### Mock测试
- **服务器模拟**: 使用mock服务器进行测试
- **网络模拟**: 模拟网络延迟和错误
- **数据模拟**: 模拟各种数据格式和错误情况

## 开发流程

### TDD循环
1. **编写测试**: 先编写失败的测试用例
2. **运行测试**: 确认测试失败
3. **实现代码**: 编写最小代码使测试通过
4. **重构**: 优化代码结构，保持测试通过
5. **重复**: 继续下一个功能

### 代码质量
- **类型安全**: 强类型系统确保编译时错误检查
- **错误处理**: 完善的错误类型和处理机制
- **文档**: 完整的API文档和使用示例
- **性能**: 异步操作和连接复用

## 部署和使用

### 构建
```bash
cargo build --release
```

### 测试
```bash
# 单元测试
cargo test

# 集成测试
cargo test --test integration_tests

# 运行示例
cargo run --example basic_usage
```

### 使用示例
```rust
use marketstore_rust_client::MarketStoreClient;

#[tokio::main]
async fn main() -> Result<()> {
    let mut client = MarketStoreClient::new(
        "http://localhost:5995".to_string(),
        "ws://localhost:5993/ws".to_string(),
    ).await?;

    // 查询数据
    let data = client.query("BTCUSDT", "1Min", "OHLCV", None, None, None).await?;
    
    // 订阅实时数据
    let subscription = StreamSubscription::new().add_stream("BTCUSDT/1Min/OHLCV");
    let handle = client.subscribe_realtime(subscription, |payload| {
        println!("Received: {:?}", payload);
        Ok(())
    }).await?;
    
    Ok(())
}
```

## 性能优化

### 连接管理
- **连接池**: gRPC连接复用
- **异步操作**: 非阻塞I/O操作
- **批量处理**: 减少网络往返次数

### 内存管理
- **零拷贝**: 最小化数据复制
- **智能指针**: 自动内存管理
- **流式处理**: 大数据的流式处理

### 并发处理
- **多线程**: 利用多核CPU
- **异步任务**: 高效的异步调度
- **锁优化**: 最小化锁竞争

## 监控和调试

### 日志系统
- **结构化日志**: 使用tracing框架
- **日志级别**: 可配置的日志级别
- **性能指标**: 连接时间、查询延迟等

### 错误追踪
- **错误分类**: 详细的错误类型
- **错误上下文**: 错误发生的位置和原因
- **错误恢复**: 自动重试和恢复机制

## 扩展计划

### 短期目标
- ✅ 基础gRPC和WebSocket支持
- ✅ 完整的CRUD操作
- ✅ 实时数据订阅
- ✅ 错误处理和重试机制

### 中期目标
- 🔄 连接池优化
- 🔄 批量操作优化
- 🔄 性能监控
- 🔄 更多数据格式支持

### 长期目标
- 📋 分布式支持
- 📋 高可用性
- 📋 插件系统
- 📋 云原生部署

## 总结

这个TDD开发计划提供了一个完整的MarketStore Rust客户端开发框架，具有以下特点：

1. **完整的TDD流程**: 从测试到实现，确保代码质量
2. **模块化设计**: 清晰的模块分离和职责划分
3. **高性能**: 异步操作和连接复用
4. **类型安全**: Rust的强类型系统
5. **易于使用**: 简洁的API设计
6. **可扩展**: 支持未来的功能扩展

通过这个计划，可以开发出一个高质量、高性能的MarketStore Rust客户端，满足量化交易和金融数据处理的严格要求。 