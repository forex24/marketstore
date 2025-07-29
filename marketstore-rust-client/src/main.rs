use marketstore_rust_client::{
    MarketStoreClient, OHLCVData, StreamSubscription, SymbolFormat, DataShape,
    error::Result,
};
use tokio::sync::oneshot;
use tracing::info;

#[tokio::main]
async fn main() -> Result<()> {
    // 初始化日志
    tracing_subscriber::fmt::init();
    
    info!("Starting MarketStore Rust Client Example");
    
    // 创建客户端
    let mut client = MarketStoreClient::new(
        "http://localhost:5995".to_string(),
        "ws://localhost:5993/ws".to_string(),
    ).await?;

    info!("Connected to MarketStore");

    // 健康检查
    let is_healthy = client.health_check().await?;
    info!("Health check: {}", is_healthy);

    // 获取服务器版本
    let version = client.server_version().await?;
    info!("Server version: {}", version);

    // 列出所有symbols
    let symbols = client.list_symbols(SymbolFormat::Symbol).await?;
    info!("Available symbols: {:?}", symbols);

    // 创建bucket (如果不存在)
    let data_shapes = vec![
        DataShape {
            name: "Epoch".to_string(),
            data_type: "i8".to_string(),
        },
        DataShape {
            name: "Open".to_string(),
            data_type: "f4".to_string(),
        },
        DataShape {
            name: "High".to_string(),
            data_type: "f4".to_string(),
        },
        DataShape {
            name: "Low".to_string(),
            data_type: "f4".to_string(),
        },
        DataShape {
            name: "Close".to_string(),
            data_type: "f4".to_string(),
        },
        DataShape {
            name: "Volume".to_string(),
            data_type: "f4".to_string(),
        },
    ];

    client.create_bucket("TEST", "1Min", "OHLCV", data_shapes).await?;
    info!("Created bucket: TEST/1Min/OHLCV");

    // 写入数据
    let data = vec![
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

    client.write("TEST", "1Min", "OHLCV", data).await?;
    info!("Wrote data to TEST/1Min/OHLCV");

    // 查询数据
    let query_result = client.query(
        "TEST",
        "1Min",
        "OHLCV",
        Some(1640995200),
        Some(1640995260),
        Some(100),
    ).await?;

    info!("Query result: {:?}", query_result);

    // 订阅实时数据
    let subscription = StreamSubscription::new()
        .add_stream("TEST/1Min/OHLCV")
        .add_stream("BTCUSDT/1Min/OHLCV");

    let received_count = std::sync::Arc::new(std::sync::atomic::AtomicUsize::new(0));
    let received_count_clone = received_count.clone();
    let handler = move |payload| {
        let count = received_count_clone.fetch_add(1, std::sync::atomic::Ordering::SeqCst);
        info!("Received real-time data #{}: {:?}", count + 1, payload);
        Ok(())
    };

    let (tx, rx) = oneshot::channel();
    
    let stream_handle = client.subscribe_realtime_with_cancel(subscription, handler, rx).await?;
    info!("Subscribed to real-time data streams");

    // 等待一段时间接收数据
    tokio::time::sleep(tokio::time::Duration::from_secs(10)).await;
    
    // 发送取消信号
    let _ = tx.send(());
    
    // 等待流处理完成
    let _ = stream_handle.await;
    
    info!("Received {} real-time messages", received_count.load(std::sync::atomic::Ordering::SeqCst));

    // 批量操作示例
    let batch_queries = vec![
        ("TEST", "1Min", "OHLCV"),
        ("BTCUSDT", "1Min", "OHLCV"),
    ];

    let batch_results = client.batch_query(batch_queries).await?;
    info!("Batch query results: {} datasets", batch_results.len());

    // 清理测试数据
    client.destroy_bucket("TEST", "1Min", "OHLCV").await?;
    info!("Destroyed bucket: TEST/1Min/OHLCV");

    info!("Example completed successfully");
    Ok(())
} 