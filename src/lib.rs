#![warn(missing_debug_implementations)]
#![warn(unreachable_pub)]
#![deny(unused_must_use)]

mod config;
mod fmt;

pub use fmt::rotate_log_file;
pub use fmt::{AnsiFormatter, Icons, Theme};

pub use config::{
    ConsoleConfig, ConsoleWriter, FileLoggingConfig, LogFilter, LogFormat, LogLevel, LogRotation,
    LoggingConfig,
};

use std::path::PathBuf;
use std::sync::{Arc, Mutex};
use tracing_subscriber::EnvFilter;
use tracing_subscriber::fmt::format::FmtSpan;
use tracing_subscriber::layer::Layered;
use tracing_subscriber::prelude::*;

pub type LogHandle = tracing_appender::non_blocking::WorkerGuard;

type FmtLayer = Box<dyn tracing_subscriber::Layer<tracing_subscriber::Registry> + Send + Sync>;
type InnerSubscriber = Layered<FmtLayer, tracing_subscriber::Registry>;
type RawReloadHandle = tracing_subscriber::reload::Handle<EnvFilter, InnerSubscriber>;

#[must_use = "dropping ReloadHandle loses the ability to change log filters at runtime"]
#[derive(Clone)]
pub struct ReloadHandle {
    raw: RawReloadHandle,
    filter: Arc<Mutex<LogFilter>>,
}

impl std::fmt::Debug for ReloadHandle {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("ReloadHandle").finish_non_exhaustive()
    }
}

impl ReloadHandle {
    pub fn reload(&self, directive: &str) -> anyhow::Result<()> {
        self.apply_directive(directive)?;
        self.store_filter(LogFilter::new(LogLevel::Custom(directive.to_string())))?;
        Ok(())
    }

    pub fn set_filter(&self, filter: LogFilter) -> anyhow::Result<()> {
        let directive = filter.as_filter_directive();
        self.apply_directive(&directive)?;
        self.store_filter(filter)?;
        Ok(())
    }

    pub fn set_level(&self, level: LogLevel) -> anyhow::Result<()> {
        self.update_filter(|filter| filter.level = level)
    }

    pub fn set_target_level(
        &self,
        target: impl Into<String>,
        level: LogLevel,
    ) -> anyhow::Result<()> {
        let target = target.into();
        self.update_filter(|filter| filter.set_target_level(target, level))
    }

    pub fn remove_target_level(&self, target: &str) -> anyhow::Result<()> {
        self.update_filter(|filter| {
            filter.remove_target_level(target);
        })
    }

    fn update_filter(&self, update: impl FnOnce(&mut LogFilter)) -> anyhow::Result<()> {
        let mut next = self.current_filter()?;
        update(&mut next);
        self.set_filter(next)
    }

    fn current_filter(&self) -> anyhow::Result<LogFilter> {
        Ok(self
            .filter
            .lock()
            .map_err(|_| anyhow::anyhow!("log filter state lock poisoned"))?
            .clone())
    }

    fn store_filter(&self, filter: LogFilter) -> anyhow::Result<()> {
        *self
            .filter
            .lock()
            .map_err(|_| anyhow::anyhow!("log filter state lock poisoned"))? = filter;
        Ok(())
    }

    fn apply_directive(&self, directive: &str) -> anyhow::Result<()> {
        let filter = EnvFilter::try_new(directive)?;
        self.raw.modify(|f| *f = filter)?;
        Ok(())
    }
}

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

#[must_use = "dropping FileLayerParts.guard will stop file logging"]
pub struct FileLayerParts {
    pub writer: tracing_appender::non_blocking::NonBlocking,
    pub guard: LogHandle,
    pub path: PathBuf,
}

impl std::fmt::Debug for FileLayerParts {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("FileLayerParts")
            .field("path", &self.path)
            .finish_non_exhaustive()
    }
}

pub fn build_file_layer(file_config: &FileLoggingConfig) -> anyhow::Result<FileLayerParts> {
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

pub fn build_reload_filter(
    level: &LogLevel,
) -> (
    tracing_subscriber::reload::Layer<EnvFilter, InnerSubscriber>,
    ReloadHandle,
) {
    let filter = EnvFilter::new(level.as_filter_directive());
    let (layer, raw_handle) = tracing_subscriber::reload::Layer::new(filter);
    (
        layer,
        ReloadHandle {
            raw: raw_handle,
            filter: Arc::new(Mutex::new(LogFilter::new(level.clone()))),
        },
    )
}

pub fn init_tracing(config: &LoggingConfig) -> anyhow::Result<TracingGuard> {
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
