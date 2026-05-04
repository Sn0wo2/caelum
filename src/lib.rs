#![warn(missing_debug_implementations)]
#![warn(unreachable_pub)]
#![deny(unused_must_use)]

mod config;
mod error;
mod fmt;
mod reload;
mod rotation;
#[cfg(any(feature = "custom-async", feature = "native-async"))]
mod writer;

pub use error::{Result, SageTraceError};
pub use fmt::{AnsiFormatter, Icons, LevelLabels, Theme};
pub use rotation::rotate_log_file;

#[cfg(feature = "custom-async")]
pub use writer::{AsyncWriter, AsyncWriterTarget, async_writer, async_writer_for};

#[cfg(any(feature = "custom-async", feature = "native-async"))]
pub use config::AsyncWriterMode;

pub use config::{
    ConsoleConfig, ConsoleWriter, FileLoggingConfig, FilterDirective, LogFilter, LogFormat,
    LogLevel, LogRotation, LoggingConfig,
};

#[cfg(feature = "file")]
use std::path::PathBuf;
use tracing_subscriber::fmt::format::FmtSpan;
use tracing_subscriber::prelude::*;

#[cfg(feature = "file")]
pub type LogHandle = tracing_appender::non_blocking::WorkerGuard;

use crate::reload::FmtLayer;

pub use crate::reload::ReloadHandle;

#[cfg(feature = "file")]
#[must_use = "dropping TracingGuard will stop file logging"]
#[derive(Debug)]
pub struct TracingGuard {
    pub worker_guard: Option<LogHandle>,
    pub log_path: Option<PathBuf>,
    pub reload_handle: ReloadHandle,
}

pub fn build_console_layer(console: &ConsoleConfig) -> FmtLayer {
    let formatter = AnsiFormatter::new()
        .with_show_path(console.show_path)
        .with_show_spans(console.show_spans);

    let formatter = match &console.time_format {
        Some(tf) => formatter.with_time_format(tf.clone()),
        None => formatter,
    };

    build_console_layer_with(console, formatter)
}

pub fn build_console_layer_with(console: &ConsoleConfig, formatter: AnsiFormatter) -> FmtLayer {
    macro_rules! with_writer {
        ($layer:expr, $console:expr) => {
            match $console.writer {
                ConsoleWriter::Stdout => $layer.with_writer(std::io::stdout).boxed(),
                ConsoleWriter::Stderr => $layer.with_writer(std::io::stderr).boxed(),
                #[cfg(any(feature = "custom-async", feature = "native-async"))]
                ConsoleWriter::AsyncStdout(mode) => match mode {
                    #[cfg(feature = "custom-async")]
                    config::AsyncWriterMode::Custom => $layer
                        .with_writer(writer::async_writer_for(writer::AsyncWriterTarget::Stdout))
                        .boxed(),
                    #[cfg(feature = "native-async")]
                    config::AsyncWriterMode::Native => $layer
                        .with_writer(writer::native_async_writer(
                            writer::AsyncWriterTarget::Stdout,
                        ))
                        .boxed(),
                },
                #[cfg(any(feature = "custom-async", feature = "native-async"))]
                ConsoleWriter::AsyncStderr(mode) => match mode {
                    #[cfg(feature = "custom-async")]
                    config::AsyncWriterMode::Custom => $layer
                        .with_writer(writer::async_writer_for(writer::AsyncWriterTarget::Stderr))
                        .boxed(),
                    #[cfg(feature = "native-async")]
                    config::AsyncWriterMode::Native => $layer
                        .with_writer(writer::native_async_writer(
                            writer::AsyncWriterTarget::Stderr,
                        ))
                        .boxed(),
                },
            }
        };
    }

    match &console.format {
        LogFormat::Pretty => {
            let layer = tracing_subscriber::fmt::Layer::default()
                .pretty()
                .with_target(true)
                .with_file(true)
                .with_line_number(true)
                .with_thread_ids(false)
                .with_thread_names(false)
                .with_ansi(console.ansi)
                .with_span_events(FmtSpan::NONE);
            with_writer!(layer, console)
        }
        LogFormat::Compact => {
            let layer = tracing_subscriber::fmt::Layer::default()
                .with_target(false)
                .with_file(false)
                .with_line_number(false)
                .with_thread_ids(false)
                .with_thread_names(false)
                .with_ansi(console.ansi)
                .with_span_events(FmtSpan::NONE)
                .event_format(formatter);
            with_writer!(layer, console)
        }
        LogFormat::Json => {
            let layer = tracing_subscriber::fmt::Layer::default()
                .json()
                .with_target(false)
                .with_file(false)
                .with_line_number(false)
                .with_current_span(false)
                .with_span_list(false)
                .flatten_event(true)
                .with_ansi(false);
            with_writer!(layer, console)
        }
    }
}

#[cfg(feature = "file")]
#[must_use = "dropping FileLayerParts.guard will stop file logging"]
pub struct FileLayerParts {
    pub writer: tracing_appender::non_blocking::NonBlocking,
    pub guard: LogHandle,
    pub path: PathBuf,
}

#[cfg(feature = "file")]
impl std::fmt::Debug for FileLayerParts {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("FileLayerParts")
            .field("path", &self.path)
            .finish_non_exhaustive()
    }
}

#[cfg(feature = "file")]
pub fn build_file_layer(file_config: &FileLoggingConfig) -> error::Result<FileLayerParts> {
    let path = &file_config.path;
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)?;
    }

    rotate_log_file(path, file_config.rotation.clone())?;

    let path = resolve_log_path(path);

    let file_appender = tracing_appender::rolling::never(
        path.parent().unwrap_or(std::path::Path::new(".")),
        path.file_name().unwrap_or_default(),
    );

    let (non_blocking, guard) = tracing_appender::non_blocking(file_appender);

    Ok(FileLayerParts {
        writer: non_blocking,
        guard,
        path,
    })
}

pub use crate::reload::build_reload_filter;

#[cfg(feature = "file")]
pub fn init_tracing(config: &LoggingConfig) -> error::Result<TracingGuard> {
    let (filter, reload_handle) = build_reload_filter(&config.level);

    let console_layer: FmtLayer = match &config.console {
        Some(console) => build_console_layer(console),
        None => tracing_subscriber::fmt::Layer::default()
            .with_writer(std::io::sink)
            .boxed(),
    };

    let subscriber = tracing_subscriber::Registry::default()
        .with(console_layer)
        .with(filter);

    let (worker_guard, log_path) = if let Some(file_config) = &config.file {
        let parts = build_file_layer(file_config)?;

        let file_layer = tracing_subscriber::fmt::layer()
            .json()
            .with_target(true)
            .with_file(true)
            .with_line_number(true)
            .with_current_span(true)
            .with_span_list(true)
            .flatten_event(true)
            .with_ansi(false)
            .with_writer(parts.writer);

        let subscriber = subscriber.with(file_layer);
        tracing::subscriber::set_global_default(subscriber)?;
        (Some(parts.guard), Some(parts.path))
    } else {
        tracing::subscriber::set_global_default(subscriber)?;
        (None, None)
    };

    Ok(TracingGuard {
        worker_guard,
        log_path,
        reload_handle,
    })
}

#[cfg(feature = "file")]
fn resolve_log_path(path: &std::path::Path) -> PathBuf {
    match std::fs::OpenOptions::new()
        .create(true)
        .append(true)
        .open(path)
    {
        Ok(_) => path.to_path_buf(),
        Err(_) => {
            let pid = std::process::id();
            let stem = path
                .file_stem()
                .and_then(|s| s.to_str())
                .unwrap_or("latest");
            let ext = path.extension().and_then(|s| s.to_str()).unwrap_or("log");
            path.with_file_name(format!("{stem}-{pid}.{ext}"))
        }
    }
}

#[cfg(test)]
mod test;
