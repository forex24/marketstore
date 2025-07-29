use marketstore_rust_client::{
    MarketStoreClient, OHLCVData, StreamSubscription, SymbolFormat, DataShape, StreamPayload,
    error::Result,
};
use tokio::sync::oneshot;
use tracing::info;

#[tokio::main]
async fn main() -> Result<()> {
    // 初始化日志
    tracing_subscriber::fmt::init();
    
    info!("MarketStore Rust Client - Basic Usage Example");
    
    // 创建客户端
    let mut client = MarketStoreClient::new(
        "http://localhost:5995".to_string(),
        "ws://localhost:5993/ws".to_string(),
    ).await?;

    info!("✅ Connected to MarketStore");

    // 1. 健康检查
    let is_healthy = client.health_check().await?;
    info!("Health check: {}", if is_healthy { "✅ OK" } else { "❌ Failed" });

    // 2. 获取服务器版本
    let version = client.server_version().await?;
    info!("Server version: {}", version);

    // 3. 列出所有symbols
    let symbols = client.list_symbols(SymbolFormat::Symbol).await?;
    info!("Available symbols: {} (showing first 5)", symbols.len());
    for (i, symbol) in symbols.iter().take(5).enumerate() {
        info!("  {}. {}", i + 1, symbol);
    }

    // 4. 创建测试bucket
    let data_shapes = vec![
        DataShape { name: "Epoch".to_string(), data_type: "i8".to_string() },
        DataShape { name: "Open".to_string(), data_type: "f4".to_string() },
        DataShape { name: "High".to_string(), data_type: "f4".to_string() },
        DataShape { name: "Low".to_string(), data_type: "f4".to_string() },
        DataShape { name: "Close".to_string(), data_type: "f4".to_string() },
        DataShape { name: "Volume".to_string(), data_type: "f4".to_string() },
    ];

    client.create_bucket("EXAMPLE", "1Min", "OHLCV", data_shapes).await?;
    info!("✅ Created bucket: EXAMPLE/1Min/OHLCV");

    // 5. 写入测试数据
    let test_data = vec![
        OHLCVData {
            epoch: 1640995200,
            open: 100.0,
            high: 101.0,
            low: 99.0,
            close: 100.5,
            volume: 1000.0,
        },
        OHLCVData {
            epoch: 1640995260,
            open: 100.5,
            high: 102.0,
            low: 100.0,
            close: 101.5,
            volume: 1500.0,
        },
    ];

    client.write("EXAMPLE", "1Min", "OHLCV", test_data).await?;
    info!("✅ Wrote 2 data points to EXAMPLE/1Min/OHLCV");

    // 6. 查询数据
    let query_result = client.query(
        "EXAMPLE",
        "1Min",
        "OHLCV",
        Some(1640995200),
        Some(1640995260),
        Some(100),
    ).await?;

    info!("✅ Query completed");
    if let Some(dataset) = &query_result.data {
        info!("  Records: {}", dataset.length);
        info!("  Columns: {:?}", dataset.column_names);
    }

    // 7. 订阅实时数据（如果服务器支持）
    info!("Subscribing to real-time data for 10 seconds...");
    let subscription = StreamSubscription::new()
        .add_stream("EXAMPLE/1Min/OHLCV");

    let received_count = std::sync::Arc::new(std::sync::atomic::AtomicUsize::new(0));
    let received_count_clone = received_count.clone();
    let handler = move |payload: StreamPayload| {
        let count = received_count_clone.fetch_add(1, std::sync::atomic::Ordering::SeqCst);
        info!("📡 Received #{}: {} = {:?}", count + 1, payload.key, payload.data);
        Ok(())
    };

    let (tx, rx) = oneshot::channel();
    let stream_handle = client.subscribe_realtime_with_cancel(subscription, handler, rx).await?;
    
    // 等待10秒
    tokio::time::sleep(tokio::time::Duration::from_secs(10)).await;
    
    // 发送取消信号
    let _ = tx.send(());
    let _ = stream_handle.await;
    
    info!("✅ Real-time subscription completed, received {} messages", received_count.load(std::sync::atomic::Ordering::SeqCst));

    // 8. 清理测试数据
    client.destroy_bucket("EXAMPLE", "1Min", "OHLCV").await?;
    info!("✅ Cleaned up test bucket");

    info!("🎉 Basic usage example completed successfully!");
    Ok(())
} 