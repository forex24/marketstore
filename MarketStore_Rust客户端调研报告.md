# MarketStore Rust客户端调研报告

## 1. 调研结果概述

经过网络搜索，发现了**2个MarketStore的Rust客户端**：

### 1.1 主要发现
1. **rsmarketstore** - 功能相对完整的Rust驱动
2. **marketstore.rs** - 社区贡献的Rust客户端

## 2. 详细分析

### 2.1 rsmarketstore

#### 2.1.1 基本信息
- **GitHub仓库**: https://github.com/ananthartha/rsmarketstore
- **Crates.io**: https://crates.io/crates/marketstore
- **星标数**: 4
- **最后更新**: 2023年9月16日
- **下载量**: 4,541次

#### 2.1.2 功能特性
```rust
// 连接MarketStore
let agent = Agent::connect(
    Uri::from_static("http://localhost:5995").into()
).await;

// 查询数据
agent
    .query(QueryParams {
        symbols: vec!["NIFTY 50".to_string()],
        timeframe: marketstore::MIN,
        attrgroup: "OHLCV".to_string(),
        ..Default::default()
    })
    .await?

// 时间框架常量
let FiveMins = 5 * marketstore::MIN;
let FifteenMin = 15 * marketstore::MIN;
let TwoHours = 2 * marketstore::HOUR;

// WebSocket流支持
let (stream, receiver) = stream::connect::<Candle>("ws://localhost:5993/ws")
    .await
    .unwrap();

stream.subscribe(vec!["NIFTY 50".into()]).await?;
receiver
    .for_each(|msg| async move {
        println!("{:#?}", msg);
    })
    .await;
```

#### 2.1.3 安装方式
```bash
cargo add rsmarketstore
```

#### 2.1.4 支持的功能
- ✅ **查询数据**: 支持查询OHLCV等时间序列数据
- ✅ **WebSocket流**: 支持实时数据流
- ✅ **时间框架**: 支持多种时间框架（分钟、小时等）
- ✅ **Serde集成**: 支持Protobuf序列化
- ❓ **写入数据**: 标记为TBD（待开发）

### 2.2 marketstore.rs

#### 2.2.1 基本信息
- **GitHub仓库**: https://github.com/marketstore-contrib/marketstore.rs
- **描述**: rust client for alpaca marketstore
- **星标数**: 0
- **状态**: 可能是早期项目或实验性项目

#### 2.2.2 项目状态
- 仓库存在但内容较少
- 可能是社区贡献的早期版本
- 功能完整性未知

## 3. 功能对比

### 3.1 功能对比表

| 功能 | rsmarketstore | marketstore.rs | Go客户端 | Python客户端 |
|------|---------------|----------------|----------|--------------|
| 查询数据 | ✅ | ❓ | ✅ | ✅ |
| 写入数据 | ❌ (TBD) | ❓ | ✅ | ✅ |
| WebSocket流 | ✅ | ❓ | ✅ | ✅ |
| ListSymbols | ❓ | ❓ | ✅ | ✅ |
| 时间框架支持 | ✅ | ❓ | ✅ | ✅ |
| 错误处理 | ❓ | ❓ | ✅ | ✅ |
| 文档完整性 | 中等 | 低 | 高 | 中等 |
| 社区活跃度 | 低 | 很低 | 高 | 中等 |

### 3.2 成熟度评估

#### rsmarketstore
- **成熟度**: 中等
- **优势**: 
  - 有基本的查询功能
  - 支持WebSocket流
  - 有Crates.io发布
- **劣势**:
  - 写入功能未实现
  - 社区活跃度低
  - 文档相对简单

#### marketstore.rs
- **成熟度**: 低
- **优势**: 
  - 社区贡献项目
- **劣势**:
  - 功能不明确
  - 文档缺乏
  - 活跃度很低

## 4. 使用建议

### 4.1 生产环境使用
**不推荐**使用Rust客户端用于生产环境，原因：
1. 功能不完整（缺少写入功能）
2. 社区支持有限
3. 文档不够详细
4. 错误处理机制不明确

### 4.2 实验性项目
**可以考虑**使用rsmarketstore进行实验性项目：
1. 基本查询功能可用
2. WebSocket流支持良好
3. 适合学习和原型开发

### 4.3 替代方案
推荐使用以下成熟的客户端：
1. **Go客户端**: 功能最完整，官方支持
2. **Python客户端**: 功能完整，易于使用
3. **直接HTTP API**: 使用msgpack协议，灵活性强

## 5. 代码示例

### 5.1 rsmarketstore使用示例

#### 5.1.1 基本查询
```rust
use rsmarketstore::{Agent, QueryParams};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // 连接MarketStore
    let agent = Agent::connect(
        Uri::from_static("http://localhost:5995").into()
    ).await;

    // 查询数据
    let result = agent
        .query(QueryParams {
            symbols: vec!["BTCUSDT".to_string()],
            timeframe: marketstore::MIN,
            attrgroup: "OHLCV".to_string(),
            ..Default::default()
        })
        .await?;

    println!("Query result: {:?}", result);
    Ok(())
}
```

#### 5.1.2 WebSocket流
```rust
use rsmarketstore::stream;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // 连接WebSocket流
    let (stream, receiver) = stream::connect::<Candle>("ws://localhost:5993/ws")
        .await
        .unwrap();

    // 订阅symbol
    stream.subscribe(vec!["BTCUSDT".into()]).await?;

    // 处理消息
    receiver
        .for_each(|msg| async move {
            match msg {
                Ok(candle) => println!("Received candle: {:?}", candle),
                Err(e) => eprintln!("Error: {:?}", e),
            }
        })
        .await;

    Ok(())
}
```

### 5.2 与Go客户端对比

#### Go客户端示例
```go
package main

import (
    "fmt"
    "github.com/alpacahq/marketstore/v4/frontend/client"
)

func main() {
    // 创建客户端
    cl, err := client.NewClient("http://localhost:5993/rpc")
    if err != nil {
        panic(err)
    }
    
    // 查询symbol列表
    symbols, err := cl.DoRPC("ListSymbols", &frontend.ListSymbolsRequest{
        Format: "symbol",
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Symbols: %v\n", symbols)
}
```

## 6. 总结

### 6.1 当前状态
- **存在Rust客户端**: 是的，有2个Rust客户端项目
- **功能完整性**: 部分功能可用，但不够完整
- **生产就绪**: 否，不建议用于生产环境
- **社区支持**: 有限，活跃度较低

### 6.2 建议
1. **学习目的**: 可以使用rsmarketstore进行学习和实验
2. **生产环境**: 推荐使用Go或Python客户端
3. **贡献机会**: 可以参与Rust客户端的开发和完善
4. **功能需求**: 如果需要完整功能，建议使用官方支持的客户端

### 6.3 未来展望
Rust客户端有潜力成为MarketStore的重要客户端之一，但需要：
1. 完善写入功能
2. 改进错误处理
3. 增加更多API支持
4. 提升文档质量
5. 增强社区活跃度 