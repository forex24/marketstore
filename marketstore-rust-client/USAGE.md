# MarketStore Rust客户端使用指南

## 快速开始

### 1. 构建项目

```bash
# 克隆项目
git clone <your-repo>
cd marketstore-rust-client

# 构建项目
cargo build --release

# 运行测试
cargo test
```

### 2. 运行示例

```bash
# 运行基本使用示例
cargo run --example basic_usage

# 运行功能测试工具
cargo run --bin marketstore_test -- --help
```

## 功能测试工具使用

`marketstore_test` 是一个命令行工具，用于测试MarketStore的所有功能。

### 基本用法

```bash
# 查看帮助
cargo run --bin marketstore_test -- --help

# 测试服务器健康状态
cargo run --bin marketstore_test -- health

# 获取服务器版本
cargo run --bin marketstore_test -- version
```

### 连接配置

```bash
# 使用自定义服务器地址
cargo run --bin marketstore_test -- \
  --grpc-url http://your-server:5995 \
  --websocket-url ws://your-server:5993/ws \
  health
```

### 数据管理

#### 列出Symbols

```bash
# 列出所有symbols（默认格式）
cargo run --bin marketstore_test -- list-symbols

# 列出symbols（TBK格式）
cargo run --bin marketstore_test -- list-symbols --format tbk
```

#### 创建Bucket

```bash
# 创建OHLCV bucket
cargo run --bin marketstore_test -- create-bucket BTCUSDT 1Min OHLCV

# 创建TICK bucket
cargo run --bin marketstore_test -- create-bucket BTCUSDT 1Sec TICK
```

#### 写入数据

```bash
# 写入10个数据点
cargo run --bin marketstore_test -- write BTCUSDT 1Min OHLCV --count 10

# 写入100个数据点
cargo run --bin marketstore_test -- write BTCUSDT 1Min OHLCV --count 100
```

#### 查询数据

```bash
# 查询所有数据
cargo run --bin marketstore_test -- query BTCUSDT 1Min OHLCV

# 查询指定时间范围的数据
cargo run --bin marketstore_test -- query BTCUSDT 1Min OHLCV \
  --start-time 1640995200 \
  --end-time 1640995260

# 限制返回记录数
cargo run --bin marketstore_test -- query BTCUSDT 1Min OHLCV --limit 50
```

### 实时数据订阅

```bash
# 订阅单个流
cargo run --bin marketstore_test -- subscribe BTCUSDT/1Min/OHLCV

# 订阅多个流
cargo run --bin marketstore_test -- subscribe \
  BTCUSDT/1Min/OHLCV \
  ETHUSDT/1Min/OHLCV \
  AAPL/1Min/OHLCV

# 订阅指定时间
cargo run --bin marketstore_test -- subscribe BTCUSDT/1Min/OHLCV --duration 60
```

### 批量测试

```bash
# 运行批量测试（默认symbols）
cargo run --bin marketstore_test -- batch-test

# 运行批量测试（自定义symbols）
cargo run --bin marketstore_test -- batch-test \
  --symbols "BTCUSDT,ETHUSDT,AAPL,GOOGL"
```

### 性能测试

```bash
# 运行性能测试（默认100次迭代）
cargo run --bin marketstore_test -- performance

# 运行性能测试（1000次迭代）
cargo run --bin marketstore_test -- performance --iterations 1000
```

## 编程接口使用

### 基本连接

```rust
use marketstore_rust_client::MarketStoreClient;

#[tokio::main]
async fn main() -> Result<()> {
    let mut client = MarketStoreClient::new(
        "http://localhost:5995".to_string(),
        "ws://localhost:5993/ws".to_string(),
    ).await?;
    
    // 使用客户端...
    Ok(())
}
```

### 数据操作

```rust
use marketstore_rust_client::{OHLCVData, DataShape, SymbolFormat};

// 创建数据形状
let data_shapes = vec![
    DataShape { name: "Epoch".to_string(), data_type: "i8".to_string() },
    DataShape { name: "Open".to_string(), data_type: "f4".to_string() },
    DataShape { name: "High".to_string(), data_type: "f4".to_string() },
    DataShape { name: "Low".to_string(), data_type: "f4".to_string() },
    DataShape { name: "Close".to_string(), data_type: "f4".to_string() },
    DataShape { name: "Volume".to_string(), data_type: "f4".to_string() },
];

// 创建bucket
client.create_bucket("BTCUSDT", "1Min", "OHLCV", data_shapes).await?;

// 写入数据
let data = vec![
    OHLCVData {
        epoch: 1640995200,
        open: 100.0,
        high: 101.0,
        low: 99.0,
        close: 100.5,
        volume: 1000.0,
    }
];
client.write("BTCUSDT", "1Min", "OHLCV", data).await?;

// 查询数据
let result = client.query(
    "BTCUSDT",
    "1Min",
    "OHLCV",
    Some(1640995200),
    Some(1640995260),
    Some(100),
).await?;
```

### 实时订阅

```rust
use marketstore_rust_client::StreamSubscription;
use tokio::sync::oneshot;

// 创建订阅
let subscription = StreamSubscription::new()
    .add_stream("BTCUSDT/1Min/OHLCV")
    .add_stream("ETHUSDT/1Min/OHLCV");

// 处理函数
let handler = |payload| {
    println!("Received: {:?}", payload);
    Ok(())
};

// 订阅（带取消）
let (tx, rx) = oneshot::channel();
let handle = client.subscribe_realtime_with_cancel(subscription, handler, rx).await?;

// 等待一段时间
tokio::time::sleep(tokio::time::Duration::from_secs(30)).await;

// 取消订阅
let _ = tx.send(());
let _ = handle.await;
```

### 批量操作

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

## 错误处理

```rust
use marketstore_rust_client::error::{MarketStoreError, Result};

async fn handle_operations() -> Result<()> {
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

## 配置选项

### 环境变量

```bash
# 设置日志级别
export RUST_LOG=info

# 设置MarketStore服务器地址
export MARKETSTORE_GRPC_URL=http://localhost:5995
export MARKETSTORE_WS_URL=ws://localhost:5993/ws
```

### 日志配置

```rust
use tracing_subscriber;

// 初始化日志
tracing_subscriber::fmt::init();

// 或者使用自定义配置
tracing_subscriber::fmt()
    .with_env_filter("marketstore_rust_client=debug")
    .init();
```

## 性能优化建议

### 1. 连接复用

客户端内部使用连接池，但建议在应用程序中复用客户端实例：

```rust
// 好的做法：复用客户端
let mut client = MarketStoreClient::new(grpc_url, ws_url).await?;

for symbol in symbols {
    client.query(symbol, "1Min", "OHLCV", None, None, None).await?;
}

// 避免：每次都创建新客户端
for symbol in symbols {
    let mut client = MarketStoreClient::new(grpc_url.clone(), ws_url.clone()).await?;
    client.query(symbol, "1Min", "OHLCV", None, None, None).await?;
}
```

### 2. 批量操作

使用批量操作减少网络往返：

```rust
// 批量查询比单个查询更高效
let queries = symbols.iter().map(|s| (*s, "1Min", "OHLCV")).collect();
let results = client.batch_query(queries).await?;
```

### 3. 异步处理

利用异步特性处理并发操作：

```rust
use futures::future::join_all;

let futures: Vec<_> = symbols.iter().map(|symbol| {
    client.query(symbol, "1Min", "OHLCV", None, None, None)
}).collect();

let results = join_all(futures).await;
```

## 故障排除

### 常见问题

1. **连接失败**
   ```bash
   # 检查服务器是否运行
   curl http://localhost:5995/health
   
   # 检查端口是否开放
   netstat -an | grep 5995
   ```

2. **权限问题**
   ```bash
   # 确保有足够的权限
   sudo chmod +x target/release/marketstore_test
   ```

3. **依赖问题**
   ```bash
   # 清理并重新构建
   cargo clean
   cargo build
   ```

### 调试模式

```bash
# 启用详细日志
RUST_LOG=debug cargo run --bin marketstore_test -- health

# 启用trace级别日志
RUST_LOG=trace cargo run --bin marketstore_test -- health
```

## 示例脚本

### 完整测试脚本

```bash
#!/bin/bash

echo "Starting MarketStore comprehensive test..."

# 1. 健康检查
echo "1. Testing health..."
cargo run --bin marketstore_test -- health

# 2. 创建测试数据
echo "2. Creating test data..."
cargo run --bin marketstore_test -- create-bucket TEST 1Min OHLCV
cargo run --bin marketstore_test -- write TEST 1Min OHLCV --count 100

# 3. 查询数据
echo "3. Querying data..."
cargo run --bin marketstore_test -- query TEST 1Min OHLCV

# 4. 性能测试
echo "4. Running performance test..."
cargo run --bin marketstore_test -- performance --iterations 1000

# 5. 清理
echo "5. Cleaning up..."
cargo run --bin marketstore_test -- destroy-bucket TEST 1Min OHLCV

echo "Test completed!"
```

这个使用指南提供了完整的MarketStore Rust客户端使用方法，包括命令行工具和编程接口的使用示例。 