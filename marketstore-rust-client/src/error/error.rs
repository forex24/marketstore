use thiserror::Error;

#[derive(Debug, Error)]
pub enum MarketStoreError {
    #[error("gRPC error: {0}")]
    Grpc(#[from] tonic::Status),
    
    #[error("WebSocket error: {0}")]
    WebSocket(String),
    
    #[error("Transport error: {0}")]
    Transport(#[from] tonic::transport::Error),
    
    #[error("Serialization error: {0}")]
    Serialization(String),
    
    #[error("Invalid data: {0}")]
    InvalidData(String),
    
    #[error("Connection error: {0}")]
    Connection(String),
    
    #[error("Timeout error: {0}")]
    Timeout(String),
    
    #[error("Protocol error: {0}")]
    Protocol(String),
}

pub type Result<T> = std::result::Result<T, MarketStoreError>;

impl From<serde_json::Error> for MarketStoreError {
    fn from(err: serde_json::Error) -> Self {
        MarketStoreError::Serialization(err.to_string())
    }
}

impl From<rmp_serde::encode::Error> for MarketStoreError {
    fn from(err: rmp_serde::encode::Error) -> Self {
        MarketStoreError::Serialization(err.to_string())
    }
}

impl From<rmp_serde::decode::Error> for MarketStoreError {
    fn from(err: rmp_serde::decode::Error) -> Self {
        MarketStoreError::Serialization(err.to_string())
    }
}

impl From<tokio_tungstenite::tungstenite::Error> for MarketStoreError {
    fn from(err: tokio_tungstenite::tungstenite::Error) -> Self {
        MarketStoreError::WebSocket(err.to_string())
    }
}

impl From<url::ParseError> for MarketStoreError {
    fn from(err: url::ParseError) -> Self {
        MarketStoreError::InvalidData(err.to_string())
    }
}

 