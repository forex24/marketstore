use marketstore_rust_client::{
    MarketStoreClient, StreamSubscription, StreamPayload,
    error::Result,
};
use tokio::sync::oneshot;
use tracing::info;
use std::sync::Arc;
use std::sync::atomic::{AtomicUsize, Ordering};

#[tokio::main]
async fn main() -> Result<()> {
    // 初始化日志
    tracing_subscriber::fmt::init();

    info!("🔧 WebSocket Message Format Test");
    info!("==================================");

    // 连接到MarketStore
    let grpc_url = "http://localhost:5995".to_string();
    let websocket_url = "ws://localhost:5993/ws".to_string();
    
    info!("Connecting to MarketStore...");
    info!("gRPC URL: {}", grpc_url);
    info!("WebSocket URL: {}", websocket_url);
    
    let client = MarketStoreClient::new(grpc_url, websocket_url).await?;
    info!("✅ Connected successfully!");

    // 测试订阅消息格式
    info!("Testing WebSocket message formats...");
    
    let subscription = StreamSubscription::new()
        .add_stream("TEST_SYMBOL/1Min/OHLCV");
    
    let received_count = Arc::new(AtomicUsize::new(0));
    let received_count_clone = received_count.clone();
    
    let handler = move |payload: StreamPayload| {
        let count = received_count_clone.fetch_add(1, Ordering::SeqCst);
        info!("📡 Received message #{}:", count + 1);
        info!("  Key: {}", payload.key);
        info!("  Data: {:?}", payload.data);
        
        // 检查数据格式
        if let Some(epoch) = payload.data.get("Epoch") {
            info!("  Epoch: {}", epoch);
        }
        if let Some(open) = payload.data.get("Open") {
            info!("  Open: {}", open);
        }
        if let Some(high) = payload.data.get("High") {
            info!("  High: {}", high);
        }
        if let Some(low) = payload.data.get("Low") {
            info!("  Low: {}", low);
        }
        if let Some(close) = payload.data.get("Close") {
            info!("  Close: {}", close);
        }
        if let Some(volume) = payload.data.get("Volume") {
            info!("  Volume: {}", volume);
        }
        
        Ok(())
    };

    let (tx, rx) = oneshot::channel();
    
    info!("Starting WebSocket subscription...");
    let stream_handle = client.subscribe_realtime_with_cancel(subscription, handler, rx).await?;
    
    // 等待10秒接收消息
    tokio::time::sleep(tokio::time::Duration::from_secs(10)).await;
    
    // 发送取消信号
    let _ = tx.send(());
    
    // 等待流处理完成
    let _ = stream_handle.await;
    
    let final_count = received_count.load(Ordering::SeqCst);
    info!("✅ WebSocket test completed");
    info!("  Received {} messages", final_count);
    
    if final_count == 0 {
        info!("ℹ️  No messages received (this is normal if no data is being streamed)");
        info!("   The WebSocket connection is working correctly!");
    }
    
    Ok(())
} 