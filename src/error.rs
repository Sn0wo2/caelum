use std::io;

#[derive(Debug, thiserror::Error)]
#[non_exhaustive]
pub enum SageTraceError {
    #[error("log filter state lock poisoned")]
    LockPoisoned,

    #[error("formatter style reload not configured: call build_reload_filter with a StyleConfig")]
    StyleNotConfigured,

    #[error("invalid filter directive: {0}")]
    InvalidDirective(#[from] tracing_subscriber::filter::ParseError),

    #[error("failed to reload filter: {0}")]
    Reload(#[from] tracing_subscriber::reload::Error),

    #[error("I/O error: {0}")]
    Io(#[from] io::Error),

    #[error("failed to set global tracing subscriber: {0}")]
    SetGlobalDefault(#[from] tracing::subscriber::SetGlobalDefaultError),
}

pub type Result<T> = std::result::Result<T, SageTraceError>;
