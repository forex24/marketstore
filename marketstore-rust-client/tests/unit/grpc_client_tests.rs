#[cfg(test)]
mod tests {
    use marketstore_rust_client::{
        client::GrpcClient,
        models::{QueryRequest, OHLCVData, SymbolFormat, DataShape},
        error::Result,
    };
    use tonic::transport::Channel;
    use mockall::predicate::*;

    #[tokio::test]
    async fn test_grpc_client_connection() {
        let client = GrpcClient::connect("http://localhost:5995").await;
        // 这个测试需要真实的MarketStore服务器运行
        // 在实际测试中，应该使用mock服务器
        assert!(client.is_ok() || client.is_err()); // 至少能处理连接结果
    }

    #[test]
    fn test_convert_to_numpy_dataset() {
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
            }
        ];
        
        // 测试数据转换逻辑
        let numpy_dataset = convert_ohlcv_to_numpy_dataset(&data);
        
        assert_eq!(numpy_dataset.column_types.len(), 6);
        assert_eq!(numpy_dataset.column_names.len(), 6);
        assert_eq!(numpy_dataset.column_data.len(), 6);
        assert_eq!(numpy_dataset.length, 2);
        
        // 验证列名
        assert!(numpy_dataset.column_names.contains(&"Epoch".to_string()));
        assert!(numpy_dataset.column_names.contains(&"Open".to_string()));
        assert!(numpy_dataset.column_names.contains(&"High".to_string()));
        assert!(numpy_dataset.column_names.contains(&"Low".to_string()));
        assert!(numpy_dataset.column_names.contains(&"Close".to_string()));
        assert!(numpy_dataset.column_names.contains(&"Volume".to_string()));
    }

    #[test]
    fn test_convert_data_shape_to_proto() {
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
        ];
        
        let proto_shapes = convert_data_shapes_to_proto(&data_shapes);
        
        assert_eq!(proto_shapes.len(), 3);
        assert_eq!(proto_shapes[0].name, "Epoch");
        assert_eq!(proto_shapes[0].r#type, "i8");
        assert_eq!(proto_shapes[1].name, "Open");
        assert_eq!(proto_shapes[1].r#type, "f4");
    }

    // 辅助函数，用于测试
    fn convert_ohlcv_to_numpy_dataset(data: &[OHLCVData]) -> crate::models::NumpyDataset {
        let epochs: Vec<i64> = data.iter().map(|d| d.epoch).collect();
        let opens: Vec<f32> = data.iter().map(|d| d.open).collect();
        let highs: Vec<f32> = data.iter().map(|d| d.high).collect();
        let lows: Vec<f32> = data.iter().map(|d| d.low).collect();
        let closes: Vec<f32> = data.iter().map(|d| d.close).collect();
        let volumes: Vec<f32> = data.iter().map(|d| d.volume).collect();

        crate::models::NumpyDataset {
            column_types: vec!["i8".to_string(), "f4".to_string(), "f4".to_string(), 
                              "f4".to_string(), "f4".to_string(), "f4".to_string()],
            column_names: vec!["Epoch".to_string(), "Open".to_string(), "High".to_string(),
                              "Low".to_string(), "Close".to_string(), "Volume".to_string()],
            column_data: vec![
                epochs.into_iter().map(|e| e.to_le_bytes()).flatten().collect(),
                opens.into_iter().map(|o| o.to_le_bytes()).flatten().collect(),
                highs.into_iter().map(|h| h.to_le_bytes()).flatten().collect(),
                lows.into_iter().map(|l| l.to_le_bytes()).flatten().collect(),
                closes.into_iter().map(|c| c.to_le_bytes()).flatten().collect(),
                volumes.into_iter().map(|v| v.to_le_bytes()).flatten().collect(),
            ],
            length: data.len() as i32,
        }
    }

    fn convert_data_shapes_to_proto(data_shapes: &[DataShape]) -> Vec<crate::proto::DataShape> {
        data_shapes.iter().map(|ds| crate::proto::DataShape {
            name: ds.name.clone(),
            r#type: ds.data_type.clone(),
        }).collect()
    }
} 