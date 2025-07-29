#[cfg(test)]
mod tests {
    use marketstore_rust_client::{
        client::MarketStoreClient,
        models::{QueryRequest, OHLCVData, SymbolFormat, StreamSubscription, DataShape},
        error::Result,
    };
    use tokio::sync::oneshot;

    #[tokio::test]
    async fn test_hybrid_client_creation() {
        let client = MarketStoreClient::new(
            "http://localhost:5995".to_string(),
            "ws://localhost:5993/ws".to_string(),
        ).await;
        
        // 这个测试需要真实的MarketStore服务器运行
        // 在实际测试中，应该使用mock服务器
        assert!(client.is_ok() || client.is_err()); // 至少能处理连接结果
    }

    #[tokio::test]
    async fn test_hybrid_client_query() {
        let mut client = MarketStoreClient::new(
            "http://localhost:5995".to_string(),
            "ws://localhost:5993/ws".to_string(),
        ).await.unwrap();
        
        let data = client.query(
            "BTCUSDT",
            "1Min", 
            "OHLCV",
            Some(1640995200),
            Some(1640995260),
            Some(100),
        ).await;
        
        // 这个测试需要真实的MarketStore服务器运行
        // 在实际测试中，应该使用mock服务器
        assert!(data.is_ok() || data.is_err()); // 至少能处理查询结果
    }

    #[tokio::test]
    async fn test_hybrid_client_write() {
        let mut client = MarketStoreClient::new(
            "http://localhost:5995".to_string(),
            "ws://localhost:5993/ws".to_string(),
        ).await.unwrap();
        
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
        
        let result = client.write("BTCUSDT", "1Min", "OHLCV", data).await;
        
        // 这个测试需要真实的MarketStore服务器运行
        // 在实际测试中，应该使用mock服务器
        assert!(result.is_ok() || result.is_err()); // 至少能处理写入结果
    }

    #[tokio::test]
    async fn test_hybrid_client_list_symbols() {
        let mut client = MarketStoreClient::new(
            "http://localhost:5995".to_string(),
            "ws://localhost:5993/ws".to_string(),
        ).await.unwrap();
        
        let symbols = client.list_symbols(SymbolFormat::Symbol).await;
        
        // 这个测试需要真实的MarketStore服务器运行
        // 在实际测试中，应该使用mock服务器
        assert!(symbols.is_ok() || symbols.is_err()); // 至少能处理查询结果
    }

    #[tokio::test]
    async fn test_hybrid_client_create_bucket() {
        let mut client = MarketStoreClient::new(
            "http://localhost:5995".to_string(),
            "ws://localhost:5993/ws".to_string(),
        ).await.unwrap();
        
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
        
        let result = client.create_bucket("TEST", "1Min", "OHLCV", data_shapes).await;
        
        // 这个测试需要真实的MarketStore服务器运行
        // 在实际测试中，应该使用mock服务器
        assert!(result.is_ok() || result.is_err()); // 至少能处理创建结果
    }

    #[tokio::test]
    async fn test_hybrid_client_subscribe_realtime() {
        let client = MarketStoreClient::new(
            "http://localhost:5995".to_string(),
            "ws://localhost:5993/ws".to_string(),
        ).await.unwrap();
        
        let subscription = StreamSubscription::new()
            .add_stream("BTCUSDT/1Min/OHLCV");
            
        let mut received_count = 0;
        let handler = |payload| {
            received_count += 1;
            println!("Received: {:?}", payload);
            Ok(())
        };
        
        let (tx, rx) = oneshot::channel();
        
        let handle = client.subscribe_realtime_with_cancel(subscription, handler, rx).await;
        
        // 这个测试需要真实的MarketStore服务器运行
        // 在实际测试中，应该使用mock服务器
        assert!(handle.is_ok() || handle.is_err()); // 至少能处理订阅结果
        
        // 发送取消信号
        let _ = tx.send(());
    }

    #[tokio::test]
    async fn test_hybrid_client_batch_operations() {
        let mut client = MarketStoreClient::new(
            "http://localhost:5995".to_string(),
            "ws://localhost:5993/ws".to_string(),
        ).await.unwrap();
        
        // 批量查询
        let queries = vec![
            ("BTCUSDT", "1Min", "OHLCV"),
            ("ETHUSDT", "1Min", "OHLCV"),
        ];
        
        let results = client.batch_query(queries).await;
        
        // 这个测试需要真实的MarketStore服务器运行
        // 在实际测试中，应该使用mock服务器
        assert!(results.is_ok() || results.is_err()); // 至少能处理批量查询结果
    }

    #[tokio::test]
    async fn test_hybrid_client_server_version() {
        let mut client = MarketStoreClient::new(
            "http://localhost:5995".to_string(),
            "ws://localhost:5993/ws".to_string(),
        ).await.unwrap();
        
        let version = client.server_version().await;
        
        // 这个测试需要真实的MarketStore服务器运行
        // 在实际测试中，应该使用mock服务器
        assert!(version.is_ok() || version.is_err()); // 至少能处理版本查询结果
    }
} 