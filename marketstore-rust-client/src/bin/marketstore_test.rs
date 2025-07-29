use marketstore_rust_client::{
    MarketStoreClient, OHLCVData, StreamSubscription, SymbolFormat, DataShape, StreamPayload,
    error::Result,
};
use tokio::sync::oneshot;
use tracing::{info, error, warn};
use clap::{App, Arg, SubCommand};
use std::time::{Duration, Instant};

#[tokio::main]
async fn main() -> Result<()> {
    // 初始化日志
    tracing_subscriber::fmt::init();
    
    let matches = App::new("MarketStore Test Client")
        .version("1.0")
        .about("Test MarketStore functionality with Rust client")
        .arg(
            Arg::with_name("grpc-url")
                .long("grpc-url")
                .value_name("URL")
                .help("gRPC server URL")
                .default_value("http://localhost:5995")
        )
        .arg(
            Arg::with_name("websocket-url")
                .long("websocket-url")
                .value_name("URL")
                .help("WebSocket server URL")
                .default_value("ws://localhost:5993/ws")
        )
        .subcommand(
            SubCommand::with_name("health")
                .about("Test server health")
        )
        .subcommand(
            SubCommand::with_name("version")
                .about("Get server version")
        )
        .subcommand(
            SubCommand::with_name("list-symbols")
                .about("List available symbols")
                .arg(
                    Arg::with_name("format")
                        .long("format")
                        .value_name("FORMAT")
                        .help("Symbol format (symbol or tbk)")
                        .default_value("symbol")
                )
        )
        .subcommand(
            SubCommand::with_name("create-bucket")
                .about("Create a new bucket")
                .arg(
                    Arg::with_name("symbol")
                        .required(true)
                        .help("Symbol name")
                )
                .arg(
                    Arg::with_name("timeframe")
                        .required(true)
                        .help("Timeframe (1Min, 1H, 1D)")
                )
                .arg(
                    Arg::with_name("attr-group")
                        .required(true)
                        .help("Attribute group (OHLCV, TICK, etc.)")
                )
        )
        .subcommand(
            SubCommand::with_name("write")
                .about("Write OHLCV data")
                .arg(
                    Arg::with_name("symbol")
                        .required(true)
                        .help("Symbol name")
                )
                .arg(
                    Arg::with_name("timeframe")
                        .required(true)
                        .help("Timeframe (1Min, 1H, 1D)")
                )
                .arg(
                    Arg::with_name("attr-group")
                        .required(true)
                        .help("Attribute group (OHLCV, TICK, etc.)")
                )
                .arg(
                    Arg::with_name("count")
                        .long("count")
                        .value_name("COUNT")
                        .help("Number of data points to generate")
                        .default_value("10")
                )
        )
        .subcommand(
            SubCommand::with_name("query")
                .about("Query data")
                .arg(
                    Arg::with_name("symbol")
                        .required(true)
                        .help("Symbol name")
                )
                .arg(
                    Arg::with_name("timeframe")
                        .required(true)
                        .help("Timeframe (1Min, 1H, 1D)")
                )
                .arg(
                    Arg::with_name("attr-group")
                        .required(true)
                        .help("Attribute group (OHLCV, TICK, etc.)")
                )
                .arg(
                    Arg::with_name("start-time")
                        .long("start-time")
                        .value_name("TIMESTAMP")
                        .help("Start time (Unix timestamp)")
                )
                .arg(
                    Arg::with_name("end-time")
                        .long("end-time")
                        .value_name("TIMESTAMP")
                        .help("End time (Unix timestamp)")
                )
                .arg(
                    Arg::with_name("limit")
                        .long("limit")
                        .value_name("LIMIT")
                        .help("Limit number of records")
                        .default_value("100")
                )
        )
        .subcommand(
            SubCommand::with_name("subscribe")
                .about("Subscribe to real-time data")
                .arg(
                    Arg::with_name("streams")
                        .required(true)
                        .multiple(true)
                        .help("Stream patterns (e.g., BTCUSDT/1Min/OHLCV)")
                )
                .arg(
                    Arg::with_name("duration")
                        .long("duration")
                        .value_name("SECONDS")
                        .help("Subscription duration in seconds")
                        .default_value("30")
                )
        )
        .subcommand(
            SubCommand::with_name("batch-test")
                .about("Run comprehensive batch test")
                .arg(
                    Arg::with_name("symbols")
                        .long("symbols")
                        .value_name("SYMBOLS")
                        .help("Comma-separated list of symbols")
                        .default_value("BTCUSDT,ETHUSDT,AAPL")
                )
        )
        .subcommand(
            SubCommand::with_name("performance")
                .about("Run performance test")
                .arg(
                    Arg::with_name("iterations")
                        .long("iterations")
                        .value_name("COUNT")
                        .help("Number of iterations")
                        .default_value("100")
                )
        )
        .get_matches();

    let grpc_url = matches.value_of("grpc-url").unwrap().to_string();
    let websocket_url = matches.value_of("websocket-url").unwrap().to_string();

    info!("Connecting to MarketStore...");
    info!("gRPC URL: {}", grpc_url);
    info!("WebSocket URL: {}", websocket_url);

    let mut client = MarketStoreClient::new(grpc_url, websocket_url).await?;
    info!("Connected successfully!");

    match matches.subcommand() {
        Some(("health", _)) => {
            test_health(&mut client).await?;
        }
        Some(("version", _)) => {
            test_version(&mut client).await?;
        }
        Some(("list-symbols", args)) => {
            let format = match args.value_of("format").unwrap() {
                "symbol" => SymbolFormat::Symbol,
                "tbk" => SymbolFormat::TimeBucketKey,
                _ => {
                    error!("Invalid format. Use 'symbol' or 'tbk'");
                    return Ok(());
                }
            };
            test_list_symbols(&mut client, format).await?;
        }
        Some(("create-bucket", args)) => {
            let symbol = args.value_of("symbol").unwrap();
            let timeframe = args.value_of("timeframe").unwrap();
            let attr_group = args.value_of("attr-group").unwrap();
            test_create_bucket(&mut client, symbol, timeframe, attr_group).await?;
        }
        Some(("write", args)) => {
            let symbol = args.value_of("symbol").unwrap();
            let timeframe = args.value_of("timeframe").unwrap();
            let attr_group = args.value_of("attr-group").unwrap();
            let count: usize = args.value_of("count").unwrap().parse().unwrap();
            test_write_data(&mut client, symbol, timeframe, attr_group, count).await?;
        }
        Some(("query", args)) => {
            let symbol = args.value_of("symbol").unwrap();
            let timeframe = args.value_of("timeframe").unwrap();
            let attr_group = args.value_of("attr-group").unwrap();
            let start_time = args.value_of("start-time").map(|s| s.parse().unwrap());
            let end_time = args.value_of("end-time").map(|s| s.parse().unwrap());
            let limit: i32 = args.value_of("limit").unwrap().parse().unwrap();
            test_query_data(&mut client, symbol, timeframe, attr_group, start_time, end_time, limit).await?;
        }
        Some(("subscribe", args)) => {
            let streams: Vec<String> = args.values_of("streams").unwrap().map(|s| s.to_string()).collect();
            let duration: u64 = args.value_of("duration").unwrap().parse().unwrap();
            test_subscribe_realtime(&client, streams, duration).await?;
        }
        Some(("batch-test", args)) => {
            let symbols = args.value_of("symbols").unwrap();
            let symbol_list: Vec<&str> = symbols.split(',').collect();
            test_batch_operations(&mut client, symbol_list).await?;
        }
        Some(("performance", args)) => {
            let iterations: usize = args.value_of("iterations").unwrap().parse().unwrap();
            test_performance(&mut client, iterations).await?;
        }
        _ => {
            info!("No subcommand specified. Use --help for usage information.");
        }
    }

    Ok(())
}

async fn test_health(client: &mut MarketStoreClient) -> Result<()> {
    info!("Testing server health...");
    let start = Instant::now();
    let is_healthy = client.health_check().await?;
    let duration = start.elapsed();
    
    if is_healthy {
        info!("✅ Server is healthy (took {:?})", duration);
    } else {
        warn!("❌ Server is not healthy (took {:?})", duration);
    }
    
    Ok(())
}

async fn test_version(client: &mut MarketStoreClient) -> Result<()> {
    info!("Getting server version...");
    let start = Instant::now();
    let version = client.server_version().await?;
    let duration = start.elapsed();
    
    info!("✅ Server version: {} (took {:?})", version, duration);
    Ok(())
}

async fn test_list_symbols(client: &mut MarketStoreClient, format: SymbolFormat) -> Result<()> {
    info!("Listing symbols with format: {:?}...", format);
    let start = Instant::now();
    let symbols = client.list_symbols(format).await?;
    let duration = start.elapsed();
    
    info!("✅ Found {} symbols (took {:?})", symbols.len(), duration);
    for (i, symbol) in symbols.iter().enumerate().take(10) {
        info!("  {}. {}", i + 1, symbol);
    }
    if symbols.len() > 10 {
        info!("  ... and {} more", symbols.len() - 10);
    }
    
    Ok(())
}

async fn test_create_bucket(client: &mut MarketStoreClient, symbol: &str, timeframe: &str, attr_group: &str) -> Result<()> {
    info!("Creating bucket: {}/{}/{}", symbol, timeframe, attr_group);
    
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

    let start = Instant::now();
    client.create_bucket(symbol, timeframe, attr_group, data_shapes).await?;
    let duration = start.elapsed();
    
    info!("✅ Bucket created successfully (took {:?})", duration);
    Ok(())
}

async fn test_write_data(client: &mut MarketStoreClient, symbol: &str, timeframe: &str, attr_group: &str, count: usize) -> Result<()> {
    info!("Writing {} data points to {}/{}/{}", count, symbol, timeframe, attr_group);
    
    let base_time = 1640995200; // 2022-01-01 00:00:00 UTC
    let mut data = Vec::new();
    
    for i in 0..count {
        let epoch = base_time + (i * 60) as i64; // 每分钟一个数据点
        data.push(OHLCVData {
            epoch,
            open: 100.0 + (i as f32 * 0.1),
            high: 101.0 + (i as f32 * 0.1),
            low: 99.0 + (i as f32 * 0.1),
            close: 100.5 + (i as f32 * 0.1),
            volume: 1000.0 + (i as f32 * 10.0),
        });
    }

    let start = Instant::now();
    client.write(symbol, timeframe, attr_group, data).await?;
    let duration = start.elapsed();
    
    info!("✅ Data written successfully (took {:?})", duration);
    Ok(())
}

async fn test_query_data(
    client: &mut MarketStoreClient,
    symbol: &str,
    timeframe: &str,
    attr_group: &str,
    start_time: Option<i64>,
    end_time: Option<i64>,
    limit: i32,
) -> Result<()> {
    info!("Querying data from {}/{}/{}", symbol, timeframe, attr_group);
    if let Some(start) = start_time {
        info!("  Start time: {}", start);
    }
    if let Some(end) = end_time {
        info!("  End time: {}", end);
    }
    info!("  Limit: {}", limit);

    let start = Instant::now();
    let result = client.query(symbol, timeframe, attr_group, start_time, end_time, Some(limit)).await?;
    let duration = start.elapsed();
    
    info!("✅ Query completed (took {:?})", duration);
    
    if let Some(dataset) = &result.data {
        info!("  Records: {}", dataset.length);
        info!("  Columns: {:?}", dataset.column_names);
        info!("  Types: {:?}", dataset.column_types);
    } else {
        info!("  No data returned");
    }
    
    Ok(())
}

async fn test_subscribe_realtime(client: &MarketStoreClient, streams: Vec<String>, duration: u64) -> Result<()> {
    info!("Subscribing to real-time streams: {:?}", streams);
    info!("Duration: {} seconds", duration);
    
    let subscription = StreamSubscription::new().add_streams(streams);
    
    let received_count = std::sync::Arc::new(std::sync::atomic::AtomicUsize::new(0));
    let received_count_clone = received_count.clone();
    let handler = move |payload: StreamPayload| {
        let count = received_count_clone.fetch_add(1, std::sync::atomic::Ordering::SeqCst);
        info!("📡 Received #{}: {} = {:?}", count + 1, payload.key, payload.data);
        Ok(())
    };

    let (tx, rx) = oneshot::channel();
    
    let start = Instant::now();
    let stream_handle = client.subscribe_realtime_with_cancel(subscription, handler, rx).await?;
    
    // 等待指定时间
    tokio::time::sleep(Duration::from_secs(duration)).await;
    
    // 发送取消信号
    let _ = tx.send(());
    
    // 等待流处理完成
    let _ = stream_handle.await;
    let total_duration = start.elapsed();
    
    let final_count = received_count.load(std::sync::atomic::Ordering::SeqCst);
    info!("✅ Subscription completed (took {:?})", total_duration);
    info!("  Received {} messages", final_count);
    if duration > 0 {
        let rate = final_count as f64 / duration as f64;
        info!("  Average rate: {:.2} messages/second", rate);
    }
    
    Ok(())
}

async fn test_batch_operations(client: &mut MarketStoreClient, symbols: Vec<&str>) -> Result<()> {
    info!("Running batch operations test with symbols: {:?}", symbols);
    
    // 批量创建buckets
    info!("Creating buckets...");
    let data_shapes = vec![
        DataShape { name: "Epoch".to_string(), data_type: "i8".to_string() },
        DataShape { name: "Open".to_string(), data_type: "f4".to_string() },
        DataShape { name: "High".to_string(), data_type: "f4".to_string() },
        DataShape { name: "Low".to_string(), data_type: "f4".to_string() },
        DataShape { name: "Close".to_string(), data_type: "f4".to_string() },
        DataShape { name: "Volume".to_string(), data_type: "f4".to_string() },
    ];
    
    for symbol in &symbols {
        client.create_bucket(symbol, "1Min", "OHLCV", data_shapes.clone()).await?;
    }
    info!("✅ Buckets created");
    
    // 批量写入数据
    info!("Writing data...");
    let base_time = 1640995200;
    for (i, symbol) in symbols.iter().enumerate() {
        let data = vec![
            OHLCVData {
                epoch: base_time + (i * 60) as i64,
                open: 100.0 + (i as f32 * 10.0),
                high: 101.0 + (i as f32 * 10.0),
                low: 99.0 + (i as f32 * 10.0),
                close: 100.5 + (i as f32 * 10.0),
                volume: 1000.0 + (i as f32 * 100.0),
            }
        ];
        client.write(symbol, "1Min", "OHLCV", data).await?;
    }
    info!("✅ Data written");
    
    // 批量查询
    info!("Querying data...");
    let queries: Vec<(&str, &str, &str)> = symbols.iter().map(|s| (*s, "1Min", "OHLCV")).collect();
    let results = client.batch_query(queries).await?;
    info!("✅ Batch query completed, got {} results", results.len());
    
    // 批量删除buckets
    info!("Cleaning up buckets...");
    for symbol in &symbols {
        client.destroy_bucket(symbol, "1Min", "OHLCV").await?;
    }
    info!("✅ Buckets cleaned up");
    
    Ok(())
}

async fn test_performance(client: &mut MarketStoreClient, iterations: usize) -> Result<()> {
    info!("Running performance test with {} iterations", iterations);
    
    // 创建测试bucket
    let data_shapes = vec![
        DataShape { name: "Epoch".to_string(), data_type: "i8".to_string() },
        DataShape { name: "Open".to_string(), data_type: "f4".to_string() },
        DataShape { name: "High".to_string(), data_type: "f4".to_string() },
        DataShape { name: "Low".to_string(), data_type: "f4".to_string() },
        DataShape { name: "Close".to_string(), data_type: "f4".to_string() },
        DataShape { name: "Volume".to_string(), data_type: "f4".to_string() },
    ];
    
    client.create_bucket("PERF_TEST", "1Min", "OHLCV", data_shapes).await?;
    
    // 写入测试数据
    let test_data = vec![
        OHLCVData {
            epoch: 1640995200,
            open: 100.0,
            high: 101.0,
            low: 99.0,
            close: 100.5,
            volume: 1000.0,
        }
    ];
    client.write("PERF_TEST", "1Min", "OHLCV", test_data.clone()).await?;
    
    // 查询性能测试
    let start = Instant::now();
    for i in 0..iterations {
        if i % 100 == 0 {
            info!("  Query iteration {}/{}", i, iterations);
        }
        client.query("PERF_TEST", "1Min", "OHLCV", None, None, Some(1)).await?;
    }
    let query_duration = start.elapsed();
    
    // 写入性能测试
    let start = Instant::now();
    for i in 0..iterations {
        if i % 100 == 0 {
            info!("  Write iteration {}/{}", i, iterations);
        }
        let mut data = test_data.clone();
        data[0].epoch = 1640995200 + (i as i64 * 60);
        client.write("PERF_TEST", "1Min", "OHLCV", data).await?;
    }
    let write_duration = start.elapsed();
    
    // 清理
    client.destroy_bucket("PERF_TEST", "1Min", "OHLCV").await?;
    
    // 输出结果
    info!("✅ Performance test completed");
    info!("  Query: {} iterations in {:?} ({:.2} ops/sec)", 
          iterations, query_duration, iterations as f64 / query_duration.as_secs_f64());
    info!("  Write: {} iterations in {:?} ({:.2} ops/sec)", 
          iterations, write_duration, iterations as f64 / write_duration.as_secs_f64());
    
    Ok(())
} 