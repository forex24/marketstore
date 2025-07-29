#[cfg(test)]
mod tests {
    use marketstore_rust_client::error::{MarketStoreError, Result};

    #[test]
    fn test_grpc_error_creation() {
        let error = MarketStoreError::Grpc(tonic::Status::new(
            tonic::Code::Unavailable,
            "Connection failed"
        ));
        assert!(matches!(error, MarketStoreError::Grpc(_)));
    }

    #[test]
    fn test_websocket_error_creation() {
        let error = MarketStoreError::WebSocket("Connection closed".to_string());
        assert!(matches!(error, MarketStoreError::WebSocket(_)));
    }

    #[test]
    fn test_transport_error_creation() {
        let error = MarketStoreError::Transport(tonic::transport::Error::from(
            std::io::Error::new(std::io::ErrorKind::ConnectionRefused, "Connection refused")
        ));
        assert!(matches!(error, MarketStoreError::Transport(_)));
    }

    #[test]
    fn test_serialization_error_creation() {
        let error = MarketStoreError::Serialization("Invalid JSON".to_string());
        assert!(matches!(error, MarketStoreError::Serialization(_)));
    }

    #[test]
    fn test_invalid_data_error_creation() {
        let error = MarketStoreError::InvalidData("Invalid symbol".to_string());
        assert!(matches!(error, MarketStoreError::InvalidData(_)));
    }

    #[test]
    fn test_connection_error_creation() {
        let error = MarketStoreError::Connection("Failed to connect".to_string());
        assert!(matches!(error, MarketStoreError::Connection(_)));
    }

    #[test]
    fn test_error_display() {
        let error = MarketStoreError::InvalidData("Invalid symbol".to_string());
        assert_eq!(error.to_string(), "Invalid data: Invalid symbol");
    }

    #[test]
    fn test_error_from_tonic_status() {
        let tonic_error = tonic::Status::new(tonic::Code::Unavailable, "Service unavailable");
        let error: MarketStoreError = tonic_error.into();
        assert!(matches!(error, MarketStoreError::Grpc(_)));
    }

    #[test]
    fn test_error_from_transport_error() {
        let transport_error = tonic::transport::Error::from(
            std::io::Error::new(std::io::ErrorKind::ConnectionRefused, "Connection refused")
        );
        let error: MarketStoreError = transport_error.into();
        assert!(matches!(error, MarketStoreError::Transport(_)));
    }
} 