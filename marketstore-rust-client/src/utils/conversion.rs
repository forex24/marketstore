use crate::models::{OHLCVData, NumpyDataset};

/// 将OHLCV数据转换为字节数组
pub fn ohlcv_to_bytes(data: &[OHLCVData]) -> Vec<Vec<u8>> {
    let mut result = Vec::new();
    
    // Epoch列
    let epochs: Vec<i64> = data.iter().map(|d| d.epoch).collect();
    result.push(epochs.into_iter().map(|e| e.to_le_bytes()).flatten().collect());
    
    // Open列
    let opens: Vec<f32> = data.iter().map(|d| d.open).collect();
    result.push(opens.into_iter().map(|o| o.to_le_bytes()).flatten().collect());
    
    // High列
    let highs: Vec<f32> = data.iter().map(|d| d.high).collect();
    result.push(highs.into_iter().map(|h| h.to_le_bytes()).flatten().collect());
    
    // Low列
    let lows: Vec<f32> = data.iter().map(|d| d.low).collect();
    result.push(lows.into_iter().map(|l| l.to_le_bytes()).flatten().collect());
    
    // Close列
    let closes: Vec<f32> = data.iter().map(|d| d.close).collect();
    result.push(closes.into_iter().map(|c| c.to_le_bytes()).flatten().collect());
    
    // Volume列
    let volumes: Vec<f32> = data.iter().map(|d| d.volume).collect();
    result.push(volumes.into_iter().map(|v| v.to_le_bytes()).flatten().collect());
    
    result
}

/// 从字节数组创建NumpyDataset
pub fn create_numpy_dataset_from_ohlcv(data: &[OHLCVData]) -> NumpyDataset {
    let column_data = ohlcv_to_bytes(data);
    
    NumpyDataset {
        column_types: vec!["i8".to_string(), "f4".to_string(), "f4".to_string(), 
                          "f4".to_string(), "f4".to_string(), "f4".to_string()],
        column_names: vec!["Epoch".to_string(), "Open".to_string(), "High".to_string(),
                          "Low".to_string(), "Close".to_string(), "Volume".to_string()],
        column_data,
        length: data.len() as i32,
    }
} 