#![warn(missing_debug_implementations)]
#![warn(unreachable_pub)]
#![deny(unused_must_use)]

mod config;
mod fmt;
mod reload;
#[cfg(any(feature = "custom-async", feature = "native-async"))]
mod writer;

pub use config::{Icons, LevelLabels, StyleConfig, Theme};
pub use fmt::AnsiFormatter;

pub use tracing::{
    Level as TracingLevel, debug, debug_span, error, error_span, info, info_span, trace,
    trace_span, warn, warn_span,
};

#[cfg(feature = "custom-async")]
pub use writer::{AsyncWriter, async_writer, async_writer_for};

#[cfg(any(feature = "custom-async", feature = "native-async"))]
pub use config::AsyncMode;

#[cfg(any(feature = "custom-async", feature = "native-async"))]
pub use writer::AsyncWriterTarget;

pub use config::{
    ConsoleConfig, FileConfig, Filter, Format, Level, LoggingConfig, Rotation, Writer,
};

use std::io;
#[cfg(feature = "file")]
use std::path::Path;
#[cfg(feature = "file")]
use std::path::PathBuf;
use tracing_subscriber::fmt::format::FmtSpan;
use tracing_subscriber::prelude::*;

#[cfg(feature = "file")]
pub type LogHandle = tracing_appender::non_blocking::WorkerGuard;

use crate::reload::FmtLayer;

#[cfg(feature = "file")]
use tracing_log::LogTracer;

pub use crate::reload::ReloadHandle;

#[derive(Debug, thiserror::Error)]
#[non_exhaustive]
pub enum ActaError {
    #[error("log filter state lock poisoned")]
    LockPoisoned,

    #[error("invalid filter directive: {0}")]
    InvalidDirective(#[from] tracing_subscriber::filter::ParseError),

    #[error("failed to reload filter: {0}")]
    Reload(#[from] tracing_subscriber::reload::Error),

    #[error("I/O error: {0}")]
    Io(#[from] io::Error),

    #[error("failed to set global tracing subscriber: {0}")]
    SetGlobalDefault(#[from] tracing::subscriber::SetGlobalDefaultError),
}

pub type Result<T> = std::result::Result<T, ActaError>;

#[cfg(feature = "file")]
#[must_use = "dropping TracingGuard will stop file logging"]
#[derive(Debug)]
pub struct TracingGuard {
    #[allow(dead_code)]
    worker_guard: Option<LogHandle>,
    log_path: Option<PathBuf>,
    reload_handle: ReloadHandle,
}

#[cfg(feature = "file")]
impl TracingGuard {
    pub fn reload_handle(&self) -> &ReloadHandle {
        &self.reload_handle
    }

    pub fn reload_handle_mut(&mut self) -> &mut ReloadHandle {
        &mut self.reload_handle
    }

    pub fn log_path(&self) -> Option<&Path> {
        self.log_path.as_deref()
    }
}

pub fn build_console_layer(console: &ConsoleConfig) -> FmtLayer {
    let mut formatter = AnsiFormatter::new()
        .with_style_config(console.style)
        .with_show_path(console.show_path)
        .with_show_spans(console.show_spans);
    if let Some(tf) = &console.time_format {
        formatter = formatter.with_time_format(tf.clone());
    }
    build_console_layer_with(console, &formatter)
}

pub fn build_console_layer_with(console: &ConsoleConfig, formatter: &AnsiFormatter) -> FmtLayer {
    macro_rules! writer {
        ($layer:expr $(,)?) => {{
            match console.writer {
                Writer::Stdout => $layer.with_writer(std::io::stdout).boxed(),
                Writer::Stderr => $layer.with_writer(std::io::stderr).boxed(),
                #[cfg(feature = "custom-async")]
                Writer::AsyncStdout(AsyncMode::Custom) => $layer
                    .with_writer(writer::async_writer_for(writer::AsyncWriterTarget::Stdout))
                    .boxed(),
                #[cfg(feature = "native-async")]
                Writer::AsyncStdout(AsyncMode::Native) => $layer
                    .with_writer(writer::native_async_writer(
                        writer::AsyncWriterTarget::Stdout,
                    ))
                    .boxed(),
                #[cfg(feature = "custom-async")]
                Writer::AsyncStderr(AsyncMode::Custom) => $layer
                    .with_writer(writer::async_writer_for(writer::AsyncWriterTarget::Stderr))
                    .boxed(),
                #[cfg(feature = "native-async")]
                Writer::AsyncStderr(AsyncMode::Native) => $layer
                    .with_writer(writer::native_async_writer(
                        writer::AsyncWriterTarget::Stderr,
                    ))
                    .boxed(),
            }
        }};
    }

    let base = || {
        tracing_subscriber::fmt::Layer::default()
            .with_thread_ids(false)
            .with_thread_names(false)
            .with_span_events(FmtSpan::NONE)
    };

    match &console.format {
        Format::Pretty => writer!(
            base()
                .pretty()
                .with_target(true)
                .with_file(true)
                .with_line_number(true)
                .with_ansi(console.ansi)
        ),
        Format::Compact => writer!(
            base()
                .with_target(false)
                .with_file(false)
                .with_line_number(false)
                .with_ansi(console.ansi)
                .event_format(formatter.clone())
        ),
        Format::Json => writer!(
            base()
                .json()
                .with_target(false)
                .with_file(false)
                .with_line_number(false)
                .with_current_span(false)
                .with_span_list(false)
                .flatten_event(true)
                .with_ansi(false)
        ),
    }
}

#[cfg(feature = "file")]
pub fn build_file_layer(
    file_config: &FileConfig,
) -> Result<(
    tracing_appender::non_blocking::NonBlocking,
    LogHandle,
    PathBuf,
)> {
    let path = &file_config.path;
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)?;
    }

    rotate_log_file(path, file_config.rotation)?;
    let path = resolve_log_path(path);

    let (non_blocking, guard) = tracing_appender::non_blocking(tracing_appender::rolling::never(
        path.parent().unwrap_or(Path::new(".")),
        path.file_name().unwrap_or_default(),
    ));

    Ok((non_blocking, guard, path))
}

pub use crate::reload::build_reload_filter;

#[cfg(feature = "file")]
pub fn init_tracing(config: &LoggingConfig) -> Result<TracingGuard> {
    let (filter, reload_handle) = build_reload_filter(
        config.level.clone(),
        config
            .console
            .as_ref()
            .map_or_else(StyleConfig::default, |c| c.style),
    );

    let subscriber = tracing_subscriber::Registry::default()
        .with(match &config.console {
            Some(console) => build_console_layer(console),
            None => tracing_subscriber::fmt::Layer::default()
                .with_writer(io::sink)
                .boxed(),
        })
        .with(filter);

    let (worker_guard, log_path) = if let Some(file_config) = &config.file {
        let (writer, guard, path) = build_file_layer(file_config)?;

        let subscriber = subscriber.with(
            tracing_subscriber::fmt::layer()
                .json()
                .with_target(true)
                .with_file(true)
                .with_line_number(true)
                .with_current_span(true)
                .with_span_list(true)
                .flatten_event(true)
                .with_ansi(false)
                .with_writer(writer),
        );

        tracing::subscriber::set_global_default(subscriber)?;
        (Some(guard), Some(path))
    } else {
        tracing::subscriber::set_global_default(subscriber)?;
        (None, None)
    };

    let _ = LogTracer::init();

    Ok(TracingGuard {
        worker_guard,
        log_path,
        reload_handle,
    })
}

/// Rotate an existing log file according to `mode`.
///
/// - `Rotation::None`: no-op.
/// - `Rotation::Rename`: rename to `<stem>.<timestamp>.log`.
/// - `Rotation::Compress`: gzip + delete original.
pub fn rotate_log_file(path: &Path, mode: Rotation) -> Result<()> {
    if !path.exists() {
        return Ok(());
    }

    match mode {
        Rotation::None => Ok(()),
        Rotation::Rename => {
            let renamed = path.with_extension(format!("{}.log", now_timestamp()));
            std::fs::rename(path, renamed)?;
            Ok(())
        }
        #[cfg(feature = "compress")]
        Rotation::Compress => {
            use std::io::Write;
            let gz_path = path.with_extension(format!("{}.log.gz", now_timestamp()));
            let input = std::fs::read(path)?;
            let output = std::fs::File::create(&gz_path)?;
            let mut encoder = flate2::write::GzEncoder::new(output, flate2::Compression::default());
            encoder.write_all(&input)?;
            encoder.finish()?;
            std::fs::remove_file(path)?;
            Ok(())
        }
    }
}

fn now_timestamp() -> String {
    chrono::Local::now().format("%Y-%m-%d_%H-%M-%S").to_string()
}

#[cfg(feature = "file")]
pub(crate) fn resolve_log_path(path: &Path) -> PathBuf {
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
