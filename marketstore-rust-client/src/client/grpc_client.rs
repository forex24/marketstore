use tonic::{transport::Channel, Request};
use async_trait::async_trait;
use crate::{
    error::{MarketStoreError, Result},
    models::{QueryRequest, OHLCVData, SymbolFormat, DataShape, NumpyMultiDataset},
};

// 生成的protobuf代码
pub mod proto {
    tonic::include_proto!("proto");
}

use proto::{
    marketstore_client::MarketstoreClient,
    MultiQueryRequest, MultiWriteRequest, MultiCreateRequest, MultiKeyRequest,
    ListSymbolsRequest, QueryRequest as ProtoQueryRequest, WriteRequest,
    NumpyMultiDataset as ProtoNumpyMultiDataset, NumpyDataset as ProtoNumpyDataset,
    DataShape as ProtoDataShape,
};

#[async_trait]
pub trait GrpcClientTrait {
    async fn query(&mut self, request: QueryRequest) -> Result<NumpyMultiDataset>;
    async fn write(&mut self, symbol: &str, timeframe: &str, attr_group: &str, data: Vec<OHLCVData>) -> Result<()>;
    async fn list_symbols(&mut self, format: SymbolFormat) -> Result<Vec<String>>;
    async fn create_bucket(&mut self, symbol: &str, timeframe: &str, attr_group: &str, data_shapes: Vec<DataShape>) -> Result<()>;
    async fn destroy_bucket(&mut self, symbol: &str, timeframe: &str, attr_group: &str) -> Result<()>;
    async fn server_version(&mut self) -> Result<String>;
}

pub struct GrpcClient {
    client: MarketstoreClient<Channel>,
}

impl GrpcClient {
    pub async fn connect(addr: String) -> Result<Self> {
        let channel = Channel::from_shared(addr)
            .map_err(|e| MarketStoreError::InvalidData(e.to_string()))?
            .connect()
            .await?;
        
        let client = MarketstoreClient::new(channel);
        Ok(Self { client })
    }
}

#[async_trait]
impl GrpcClientTrait for GrpcClient {
    async fn query(&mut self, request: QueryRequest) -> Result<NumpyMultiDataset> {
        let proto_request = MultiQueryRequest {
            requests: vec![ProtoQueryRequest {
                destination: request.destination,
                epoch_start: request.epoch_start.unwrap_or(0),
                epoch_end: request.epoch_end.unwrap_or(i64::MAX),
                limit_record_count: request.limit_record_count.unwrap_or(1000),
                limit_from_start: request.limit_from_start,
                columns: request.columns,
                ..Default::default()
            }],
        };

        let response = self.client
            .query(Request::new(proto_request))
            .await?;

        let proto_response = response.into_inner();
        if proto_response.responses.is_empty() {
            return Err(MarketStoreError::InvalidData("Empty response".to_string()));
        }

        let proto_dataset = &proto_response.responses[0].result;
        if let Some(dataset) = proto_dataset {
            Ok(convert_proto_to_numpy_multi_dataset(dataset))
        } else {
            Err(MarketStoreError::InvalidData("Empty dataset".to_string()))
        }
    }

    async fn write(&mut self, symbol: &str, timeframe: &str, attr_group: &str, data: Vec<OHLCVData>) -> Result<()> {
        let numpy_dataset = convert_ohlcv_to_numpy_dataset(&data);
        let key = format!("{}/{}/{}", symbol, timeframe, attr_group);
        
        let proto_dataset = ProtoNumpyMultiDataset {
            data: Some(convert_numpy_dataset_to_proto(&numpy_dataset)),
            start_index: [(key.clone(), 0)].into_iter().collect(),
            lengths: [(key, data.len() as i32)].into_iter().collect(),
        };
        
        let request = MultiWriteRequest {
            requests: vec![WriteRequest {
                data: Some(proto_dataset),
                is_variable_length: false,
            }],
        };

        self.client
            .write(Request::new(request))
            .await?;

        Ok(())
    }

    async fn list_symbols(&mut self, format: SymbolFormat) -> Result<Vec<String>> {
        let request = ListSymbolsRequest {
            format: format.into(),
        };

        let response = self.client
            .list_symbols(Request::new(request))
            .await?;

        Ok(response.into_inner().results)
    }

    async fn create_bucket(&mut self, symbol: &str, timeframe: &str, attr_group: &str, data_shapes: Vec<DataShape>) -> Result<()> {
        let key = format!("{}/{}/{}", symbol, timeframe, attr_group);
        let proto_shapes: Vec<ProtoDataShape> = data_shapes.into_iter().map(|ds| ProtoDataShape {
            name: ds.name,
            r#type: ds.data_type,
        }).collect();

        let request = MultiCreateRequest {
            requests: vec![proto::CreateRequest {
                key,
                data_shapes: proto_shapes,
                row_type: "fixed".to_string(),
            }],
        };

        self.client
            .create(Request::new(request))
            .await?;

        Ok(())
    }

    async fn destroy_bucket(&mut self, symbol: &str, timeframe: &str, attr_group: &str) -> Result<()> {
        let key = format!("{}/{}/{}", symbol, timeframe, attr_group);
        
        let request = MultiKeyRequest {
            requests: vec![proto::KeyRequest { key }],
        };

        self.client
            .destroy(Request::new(request))
            .await?;

        Ok(())
    }

    async fn server_version(&mut self) -> Result<String> {
        let request = proto::ServerVersionRequest {};
        
        let response = self.client
            .server_version(Request::new(request))
            .await?;

        Ok(response.into_inner().version)
    }
}

// 数据转换函数
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

fn convert_numpy_dataset_to_proto(dataset: &crate::models::NumpyDataset) -> ProtoNumpyDataset {
    ProtoNumpyDataset {
        column_types: dataset.column_types.clone(),
        column_names: dataset.column_names.clone(),
        column_data: dataset.column_data.clone(),
        length: dataset.length,
        data_shapes: vec![], // 如果需要可以转换
    }
}

fn convert_proto_to_numpy_multi_dataset(proto_dataset: &ProtoNumpyMultiDataset) -> NumpyMultiDataset {
    let data = if let Some(proto_data) = &proto_dataset.data {
        Some(crate::models::NumpyDataset {
            column_types: proto_data.column_types.clone(),
            column_names: proto_data.column_names.clone(),
            column_data: proto_data.column_data.clone(),
            length: proto_data.length,
        })
    } else {
        None
    };

    NumpyMultiDataset {
        data,
        start_index: proto_dataset.start_index.clone(),
        lengths: proto_dataset.lengths.clone(),
    }
} 