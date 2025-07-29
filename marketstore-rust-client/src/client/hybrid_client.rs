use std::sync::Arc;
use tokio::sync::Mutex;
use tokio::sync::oneshot;
use crate::{
    error::Result,
    models::{QueryRequest, OHLCVData, StreamSubscription, SymbolFormat, DataShape, NumpyMultiDataset},
    client::{GrpcClient, WebSocketClient, GrpcClientTrait},
};

pub struct MarketStoreClient {
    grpc_client: Arc<Mutex<GrpcClient>>,
    websocket_url: String,
}

impl MarketStoreClient {
    pub async fn new(grpc_url: String, websocket_url: String) -> Result<Self> {
        let grpc_client = GrpcClient::connect(grpc_url).await?;
        
        Ok(Self {
            grpc_client: Arc::new(Mutex::new(grpc_client)),
            websocket_url,
        })
    }
    
    pub async fn query(
        &mut self,
        symbol: &str,
        timeframe: &str,
        attr_group: &str,
        start_time: Option<i64>,
        end_time: Option<i64>,
        limit: Option<i32>,
    ) -> Result<NumpyMultiDataset> {
        let request = QueryRequest::builder()
            .symbol(symbol)
            .timeframe(timeframe)
            .attr_group(attr_group)
            .start_time(start_time.unwrap_or(0))
            .end_time(end_time.unwrap_or(i64::MAX))
            .limit(limit.unwrap_or(1000))
            .build()?;
            
        let mut client = self.grpc_client.lock().await;
        client.query(request).await
    }
    
    pub async fn write(
        &mut self,
        symbol: &str,
        timeframe: &str,
        attr_group: &str,
        data: Vec<OHLCVData>,
    ) -> Result<()> {
        let mut client = self.grpc_client.lock().await;
        client.write(symbol, timeframe, attr_group, data).await
    }
    
    pub async fn list_symbols(&mut self, format: SymbolFormat) -> Result<Vec<String>> {
        let mut client = self.grpc_client.lock().await;
        client.list_symbols(format).await
    }
    
    pub async fn create_bucket(
        &mut self,
        symbol: &str,
        timeframe: &str,
        attr_group: &str,
        data_shapes: Vec<DataShape>,
    ) -> Result<()> {
        let mut client = self.grpc_client.lock().await;
        client.create_bucket(symbol, timeframe, attr_group, data_shapes).await
    }
    
    pub async fn destroy_bucket(
        &mut self,
        symbol: &str,
        timeframe: &str,
        attr_group: &str,
    ) -> Result<()> {
        let mut client = self.grpc_client.lock().await;
        client.destroy_bucket(symbol, timeframe, attr_group).await
    }
    
    pub async fn server_version(&mut self) -> Result<String> {
        let mut client = self.grpc_client.lock().await;
        client.server_version().await
    }
    
    pub async fn subscribe_realtime<F>(
        &self,
        subscription: StreamSubscription,
        handler: F,
    ) -> Result<tokio::task::JoinHandle<Result<()>>>
    where
        F: FnMut(crate::models::StreamPayload) -> Result<()> + Send + 'static,
    {
        let websocket_url = self.websocket_url.clone();
        
        let handle = tokio::spawn(async move {
            let ws_client = WebSocketClient::connect(&websocket_url).await?;
            ws_client.subscribe_with_handler(subscription, handler).await
        });
        
        Ok(handle)
    }
    
    pub async fn subscribe_realtime_with_cancel<F>(
        &self,
        subscription: StreamSubscription,
        handler: F,
        cancel: oneshot::Receiver<()>,
    ) -> Result<tokio::task::JoinHandle<Result<()>>>
    where
        F: FnMut(crate::models::StreamPayload) -> Result<()> + Send + 'static,
    {
        let websocket_url = self.websocket_url.clone();
        
        let handle = tokio::spawn(async move {
            let ws_client = WebSocketClient::connect(&websocket_url).await?;
            ws_client.subscribe_with_handler_and_cancel(subscription, handler, cancel).await
        });
        
        Ok(handle)
    }
    
    pub async fn batch_query(
        &mut self,
        queries: Vec<(&str, &str, &str)>,
    ) -> Result<Vec<NumpyMultiDataset>> {
        let mut client = self.grpc_client.lock().await;
        let mut results = Vec::new();
        
        for (symbol, timeframe, attr_group) in queries {
            let request = QueryRequest::builder()
                .symbol(symbol)
                .timeframe(timeframe)
                .attr_group(attr_group)
                .build()?;
                
            let result = client.query(request).await?;
            results.push(result);
        }
        
        Ok(results)
    }
    
    pub async fn batch_write(
        &mut self,
        writes: Vec<(&str, &str, &str, Vec<OHLCVData>)>,
    ) -> Result<()> {
        let mut client = self.grpc_client.lock().await;
        
        for (symbol, timeframe, attr_group, data) in writes {
            client.write(symbol, timeframe, attr_group, data).await?;
        }
        
        Ok(())
    }
    
    pub async fn health_check(&mut self) -> Result<bool> {
        match self.server_version().await {
            Ok(_) => Ok(true),
            Err(_) => Ok(false),
        }
    }
} 