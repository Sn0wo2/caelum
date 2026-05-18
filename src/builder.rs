use std::io;
#[cfg(feature = "file")]
use std::path::PathBuf;
use tracing_subscriber::fmt::format::FmtSpan;
use tracing_subscriber::prelude::*;

#[cfg(any(feature = "custom-async", feature = "native-async"))]
use crate::config::AsyncMode;
use crate::config::{ColorDepth, Config, Console, Filter, Format, Style, Writer};
use crate::fmt::Formatter;
use crate::reload::{FmtLayer, InnerSubscriber, ReloadHandle};

#[cfg(any(feature = "file", feature = "custom-async", feature = "native-async"))]
use crate::writer;

pub fn build_console_layer(console: &Console) -> FmtLayer {
    let mut formatter = Formatter::new()
        .with_style_config(console.style)
        .with_show_path(console.show_path)
        .with_show_spans(console.show_spans)
        .with_color_depth(console.color_depth.unwrap_or(ColorDepth::TrueColor));
    if let Some(tf) = &console.time_format {
        formatter = formatter.with_time_format(tf);
    }
    build_console_layer_with(console, formatter)
}

pub fn build_console_layer_with(console: &Console, formatter: Formatter) -> FmtLayer {
    macro_rules! writer {
        ($layer:expr $(,)?) => {{
            match console.writer {
                Writer::Stdout => $layer.with_writer(io::stdout).boxed(),
                Writer::Stderr => $layer.with_writer(io::stderr).boxed(),
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
        Format::Pretty(cfg) => writer!(
            base()
                .pretty()
                .with_target(cfg.target)
                .with_file(cfg.file)
                .with_line_number(cfg.line_number)
                .with_ansi(console.ansi)
        ),
        Format::Compact(cfg) => writer!(
            base()
                .with_target(cfg.target)
                .with_file(cfg.file)
                .with_line_number(cfg.line_number)
                .with_ansi(console.ansi)
                .event_format(formatter)
        ),
        Format::Json(cfg) => writer!(
            base()
                .json()
                .with_target(cfg.target)
                .with_file(cfg.file)
                .with_line_number(cfg.line_number)
                .with_current_span(cfg.current_span)
                .with_span_list(cfg.span_list)
                .flatten_event(cfg.flatten_event)
                .with_ansi(console.ansi)
        ),
    }
}

#[cfg(feature = "file")]
pub use crate::writer::build_file_layer;

pub fn build_reload_filter(
    level: crate::config::Level,
    style: Style,
) -> (
    tracing_subscriber::reload::Layer<tracing_subscriber::EnvFilter, InnerSubscriber>,
    ReloadHandle,
) {
    let filter = Filter::new(level);
    let (layer, raw) = tracing_subscriber::reload::Layer::new(
        tracing_subscriber::EnvFilter::try_new(filter.as_directive()).unwrap_or_default(),
    );
    let reload_handle = ReloadHandle { raw, filter, style };
    (layer, reload_handle)
}

#[cfg(feature = "file")]
#[non_exhaustive]
#[derive(Debug)]
pub struct SubscriberParts {
    reload_handle: ReloadHandle,
    file_writer: Option<writer::FileWriter>,
    file_guard: Option<writer::LogHandle>,
    log_path: Option<PathBuf>,
}

#[cfg(not(feature = "file"))]
#[derive(Debug)]
pub struct SubscriberParts {
    reload_handle: ReloadHandle,
}
impl SubscriberParts {
    pub fn reload_handle(self) -> ReloadHandle {
        self.reload_handle
    }

    #[cfg(feature = "file")]
    pub fn file_writer(self) -> Option<writer::FileWriter> {
        self.file_writer
    }

    #[cfg(feature = "file")]
    pub fn file_guard(self) -> Option<writer::LogHandle> {
        self.file_guard
    }

    #[cfg(feature = "file")]
    pub fn log_path(self) -> Option<PathBuf> {
        self.log_path
    }
}

#[cfg(feature = "file")]
#[allow(clippy::type_complexity)]
pub fn build_subscriber(config: &Config) -> crate::Result<SubscriberParts> {
    let console_layer = config.console.as_ref().map_or_else(
        || {
            tracing_subscriber::fmt::Layer::default()
                .with_writer(io::sink)
                .boxed()
        },
        build_console_layer,
    );

    let subscriber_with_console = tracing_subscriber::Registry::default().with(console_layer);

    let filter = Filter::new(config.level.clone());
    let env_filter =
        tracing_subscriber::EnvFilter::try_new(filter.as_directive()).unwrap_or_default();

    let (env_filter_layer, env_filter_handle) = tracing_subscriber::reload::Layer::new(env_filter);

    let subscriber = subscriber_with_console.with(env_filter_layer);

    let reload_handle = ReloadHandle {
        raw: env_filter_handle,
        filter,
        style: config
            .console
            .as_ref()
            .map_or_else(Style::default, |c| c.style),
    };

    let subscriber = if let Some(file_config) = &config.file {
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
                .with_writer(writer.clone()),
        );
        let _ = tracing_log::LogTracer::init();
        tracing::subscriber::set_global_default(subscriber)?;
        return Ok(SubscriberParts {
            reload_handle,
            file_writer: Some(writer),
            file_guard: Some(guard),
            log_path: Some(path),
        });
    } else {
        subscriber
    };

    let _ = tracing_log::LogTracer::init();

    tracing::subscriber::set_global_default(subscriber)?;
    Ok(SubscriberParts {
        reload_handle,
        #[cfg(feature = "file")]
        file_writer: None,
        #[cfg(feature = "file")]
        file_guard: None,
        #[cfg(feature = "file")]
        log_path: None,
    })
}

#[cfg(not(feature = "file"))]
pub fn build_subscriber(config: &Config) -> crate::Result<SubscriberParts> {
    let console_layer = config.console.as_ref().map_or_else(
        || {
            tracing_subscriber::fmt::Layer::default()
                .with_writer(io::sink)
                .boxed()
        },
        build_console_layer,
    );

    let subscriber_with_console = tracing_subscriber::Registry::default().with(console_layer);

    let filter = Filter::new(config.level.clone());
    let env_filter =
        tracing_subscriber::EnvFilter::try_new(filter.as_directive()).unwrap_or_default();

    let (env_filter_layer, env_filter_handle) = tracing_subscriber::reload::Layer::new(env_filter);

    let subscriber = subscriber_with_console.with(env_filter_layer);

    let reload_handle = ReloadHandle {
        raw: env_filter_handle,
        filter,
        style: config
            .console
            .as_ref()
            .map_or_else(Style::default, |c| c.style),
    };

    tracing::subscriber::set_global_default(subscriber)?;
    Ok(SubscriberParts { reload_handle })
}

#[cfg(feature = "file")]
pub fn init(config: impl Into<Config>) -> crate::Result<crate::guard::TracingGuard> {
    let SubscriberParts {
        reload_handle,
        file_guard,
        log_path,
        ..
    } = build_subscriber(&config.into())?;

    Ok(crate::guard::TracingGuard {
        reload_handle,
        #[cfg(feature = "file")]
        worker_guard: file_guard,
        #[cfg(feature = "file")]
        log_path,
    })
}

#[cfg(not(feature = "file"))]
pub fn init(config: impl Into<Config>) -> crate::Result<crate::guard::TracingGuard> {
    let SubscriberParts { reload_handle } = build_subscriber(&config.into())?;

    Ok(crate::guard::TracingGuard { reload_handle })
}
