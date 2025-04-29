use thiserror::Error;

#[derive(Error, Debug)]
pub enum GatewayError {
    #[error("Missing variable {0}")]
    MissingVariable(String),
    #[error(transparent)]
    StdIOError(#[from] std::io::Error),
    #[error(transparent)]
    ParseError(#[from] serde_json::Error),
    #[error("Error decoding argument: {0}")]
    DecodeError(#[from] base64::DecodeError),
    #[error("Custom Error: {0}")]
    CustomError(String),
    #[error("Function get is not implemented")]
    FunctionGetNotImplemented,
    #[error(transparent)]
    ModelError(#[from] crate::model::error::ModelError),
    #[error("Tool call id not found in request")]
    ToolCallIdNotFound,
    #[error(transparent)]
    ReqwestError(#[from] reqwest::Error),
    #[error(transparent)]
    BoxedError(#[from] Box<dyn std::error::Error + Send + Sync>),
}
