# MarketStore Rust Client

一个高性能的MarketStore Rust客户端，支持gRPC和WebSocket协议。

## 特性

- ✅ **gRPC支持**: 完整的CRUD操作（查询、写入、创建、删除）
- ✅ **WebSocket支持**: 实时数据流订阅
- ✅ **混合模式**: gRPC + WebSocket的最佳组合
- ✅ **异步支持**: 基于tokio的高性能异步运行时
- ✅ **类型安全**: 强类型的数据模型和错误处理
- ✅ **TDD开发**: 完整的测试覆盖
- ✅ **批量操作**: 支持批量查询和写入
- ✅ **健康检查**: 内置连接健康检查
- ✅ **取消支持**: 支持优雅的流取消

## 快速开始

### 安装

```bash
git clone <your-repo>
cd marketstore-rust-client
cargo build
```

### 基本使用

```rust
use marketstore_rust_client::{
    MarketStoreClient, OHLCVData, StreamSubscription, SymbolFormat,
    error::Result,
};

#[tokio::main]
async fn main() -> Result<()> {
    // 创建客户端
    let mut client = MarketStoreClient::new(
        "http://localhost:5995".to_string(),
        "ws://localhost:5993/ws".to_string(),
    ).await?;

    // 查询数据
    let data = client.query(
        "BTCUSDT",
        "1Min",
        "OHLCV",
        Some(1640995200),
        Some(1640995260),
        Some(100),
    ).await?;

    // 写入数据
    let ohlcv_data = vec![
        OHLCVData {
            epoch: 1640995200,
            open: 100.0,
            high: 101.0,
            low: 99.0,
            close: 100.5,
            volume: 1000.0,
        }
    ];
    
    client.write("BTCUSDT", "1Min", "OHLCV", ohlcv_data).await?;

    // 订阅实时数据
    let subscription = StreamSubscription::new()
        .add_stream("BTCUSDT/1Min/OHLCV");
        
    let handler = |payload| {
        println!("Received: {:?}", payload);
        Ok(())
    };
    
    let handle = client.subscribe_realtime(subscription, handler).await?;
    
    Ok(())
}
```

## API文档

### 核心客户端

#### MarketStoreClient

主要的客户端类，提供gRPC和WebSocket的混合功能。

```rust
impl MarketStoreClient {
    // 创建新客户端
    pub async fn new(grpc_url: String, websocket_url: String) -> Result<Self>
    
    // 查询数据
    pub async fn query(
        &mut self,
        symbol: &str,
        timeframe: &str,
        attr_group: &str,
        start_time: Option<i64>,
        end_time: Option<i64>,
        limit: Option<i32>,
    ) -> Result<NumpyMultiDataset>
    
    // 写入数据
    pub async fn write(
        &mut self,
        symbol: &str,
        timeframe: &str,
        attr_group: &str,
        data: Vec<OHLCVData>,
    ) -> Result<()>
    
    // 列出symbols
    pub async fn list_symbols(&mut self, format: SymbolFormat) -> Result<Vec<String>>
    
    // 创建bucket
    pub async fn create_bucket(
        &mut self,
        symbol: &str,
        timeframe: &str,
        attr_group: &str,
        data_shapes: Vec<DataShape>,
    ) -> Result<()>
    
    // 删除bucket
    pub async fn destroy_bucket(
        &mut self,
        symbol: &str,
        timeframe: &str,
        attr_group: &str,
    ) -> Result<()>
    
    // 订阅实时数据
    pub async fn subscribe_realtime<F>(
        &self,
        subscription: StreamSubscription,
        handler: F,
    ) -> Result<JoinHandle<Result<()>>>
    
    // 批量查询
    pub async fn batch_query(
        &mut self,
        queries: Vec<(&str, &str, &str)>,
    ) -> Result<Vec<NumpyMultiDataset>>
    
    // 健康检查
    pub async fn health_check(&mut self) -> Result<bool>
}
```

### 数据模型

#### OHLCVData

OHLCV数据结构：

```rust
pub struct OHLCVData {
    pub epoch: i64,
    pub open: f32,
    pub high: f32,
    pub low: f32,
    pub close: f32,
    pub volume: f32,
}
```

#### StreamSubscription

实时数据订阅：

```rust
pub struct StreamSubscription {
    pub streams: Vec<String>,
}

impl StreamSubscription {
    pub fn new() -> Self
    pub fn add_stream(self, stream: &str) -> Self
    pub fn add_streams(self, streams: Vec<String>) -> Self
}
```

### 错误处理

```rust
pub enum MarketStoreError {
    Grpc(tonic::Status),
    WebSocket(String),
    Transport(tonic::transport::Error),
    Serialization(String),
    InvalidData(String),
    Connection(String),
    Timeout(String),
    Protocol(String),
}
```

## 测试

### 运行单元测试

```bash
cargo test
```

### 运行集成测试

```bash
cargo test --test integration_tests
```

### 运行示例

```bash
cargo run --example basic_usage
```

## 配置

### MarketStore服务器配置

确保MarketStore服务器已启动并配置了以下端口：

- **gRPC端口**: 5995 (默认)
- **WebSocket端口**: 5993 (默认)

### 客户端配置

```rust
// 自定义配置
let mut client = MarketStoreClient::new(
    "http://your-marketstore-host:5995".to_string(),
    "ws://your-marketstore-host:5993/ws".to_string(),
).await?;
```

## 性能优化

### 批量操作

使用批量操作来提高性能：

```rust
// 批量查询
let queries = vec![
    ("BTCUSDT", "1Min", "OHLCV"),
    ("ETHUSDT", "1Min", "OHLCV"),
    ("AAPL", "1Min", "OHLCV"),
];
let results = client.batch_query(queries).await?;

// 批量写入
let writes = vec![
    ("BTCUSDT", "1Min", "OHLCV", btc_data),
    ("ETHUSDT", "1Min", "OHLCV", eth_data),
];
client.batch_write(writes).await?;
```

### 连接池

客户端内部使用连接池来复用gRPC连接，提高性能。

## 错误处理最佳实践

```rust
use marketstore_rust_client::error::{MarketStoreError, Result};

async fn handle_marketstore_operations() -> Result<()> {
    let mut client = MarketStoreClient::new(grpc_url, ws_url).await?;
    
    match client.query("BTCUSDT", "1Min", "OHLCV", None, None, None).await {
        Ok(data) => {
            println!("Query successful: {:?}", data);
        }
        Err(MarketStoreError::Connection(msg)) => {
            eprintln!("Connection error: {}", msg);
            // 重试逻辑
        }
        Err(MarketStoreError::InvalidData(msg)) => {
            eprintln!("Invalid data: {}", msg);
            // 数据验证逻辑
        }
        Err(e) => {
            eprintln!("Unexpected error: {:?}", e);
        }
    }
    
    Ok(())
}
```

## 贡献

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打开 Pull Request

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 支持

如有问题或建议，请提交 [Issue](https://github.com/yourusername/marketstore-rust-client/issues)。 