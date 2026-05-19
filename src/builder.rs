#[cfg(any(feature = "custom-async", feature = "native-async"))]
use crate::config::AsyncMode;
use crate::config::{ColorDepth, Config, Filter, Format, Style, Writer, WriterTarget};
use crate::fmt::Formatter;
use std::io;
#[cfg(feature = "file")]
use std::path::PathBuf;
use std::sync::Arc;
use tracing_subscriber::fmt::format::FmtSpan;
use tracing_subscriber::prelude::*;

pub type FmtLayer<S = tracing_subscriber::Registry> =
    Box<dyn tracing_subscriber::Layer<S> + Send + Sync>;

// tracing_subscriber::reload::Handle<EnvFilter, S> needs the concrete S.
// Six optional layers stacked → this pyramid.  Users never name it.
type Layer1 = tracing_subscriber::layer::Layered<
    Option<FmtLayer<tracing_subscriber::Registry>>,
    tracing_subscriber::Registry,
>;
type Layer2 = tracing_subscriber::layer::Layered<Option<FmtLayer<Layer1>>, Layer1>;
type Layer3 = tracing_subscriber::layer::Layered<Option<FmtLayer<Layer2>>, Layer2>;
type Layer4 = tracing_subscriber::layer::Layered<Option<FmtLayer<Layer3>>, Layer3>;
pub(crate) type InnerSubscriber =
    tracing_subscriber::layer::Layered<Option<FmtLayer<Layer4>>, Layer4>;

pub(crate) struct ReloadHandle(
    pub(crate) tracing_subscriber::reload::Handle<tracing_subscriber::EnvFilter, InnerSubscriber>,
);

impl ReloadHandle {
    fn new(
        filter: tracing_subscriber::EnvFilter,
    ) -> (
        tracing_subscriber::reload::Layer<tracing_subscriber::EnvFilter, InnerSubscriber>,
        Self,
    ) {
        let (layer, handle) = tracing_subscriber::reload::Layer::new(filter);
        (layer, Self(handle))
    }

    fn modify(&self, f: impl FnOnce(&mut tracing_subscriber::EnvFilter)) -> crate::Result<()> {
        self.0.modify(f)?;
        Ok(())
    }
}

impl From<Style> for Arc<arc_swap::ArcSwap<Style>> {
    fn from(s: Style) -> Self {
        Self::new(arc_swap::ArcSwap::new(Arc::new(s)))
    }
}

#[cfg(any(feature = "file", feature = "custom-async", feature = "native-async"))]
use crate::writer;

fn build_formatter(
    writer: &Writer,
    color_depth: ColorDepth,
    shared_handle: Option<Arc<arc_swap::ArcSwap<Style>>>,
) -> Formatter {
    let mut f = shared_handle
        .as_ref()
        .map_or_else(Formatter::new, |handle| {
            Formatter::new_with_handle(handle.clone())
        });
    f = f
        .with_style_config(writer.style)
        .with_show_path(writer.show_path)
        .with_show_spans(writer.show_spans)
        .with_color_depth(color_depth);
    if let Some(tf) = &writer.time_format {
        f = f.with_time_format(tf);
    }
    f
}

fn build_layer_inner<S>(
    writer: &Writer,
    shared_handle: Option<Arc<arc_swap::ArcSwap<Style>>>,
) -> FmtLayer<S>
where
    S: tracing::Subscriber + for<'a> tracing_subscriber::registry::LookupSpan<'a>,
{
    let formatter = build_formatter(
        writer,
        writer.color_depth.unwrap_or_else(|| {
            if writer.ansi {
                crate::config::depth::detect(&writer.target)
            } else {
                ColorDepth::NoColor
            }
        }),
        shared_handle,
    );

    let base = || {
        tracing_subscriber::fmt::Layer::default()
            .with_thread_ids(false)
            .with_thread_names(false)
            .with_span_events(FmtSpan::NONE)
    };

    macro_rules! write_to {
        ($layer:expr $(,)?) => {{
            match &writer.target {
                WriterTarget::Stdout => $layer.with_writer(io::stdout).boxed(),
                WriterTarget::Stderr => $layer.with_writer(io::stderr).boxed(),
                #[cfg(feature = "custom-async")]
                WriterTarget::AsyncStdout(AsyncMode::Custom) => $layer
                    .with_writer(writer::async_writer_for(writer::AsyncWriterTarget::Stdout))
                    .boxed(),
                #[cfg(feature = "native-async")]
                WriterTarget::AsyncStdout(AsyncMode::Native) => $layer
                    .with_writer(writer::native_async_writer(
                        writer::AsyncWriterTarget::Stdout,
                    ))
                    .boxed(),
                #[cfg(feature = "custom-async")]
                WriterTarget::AsyncStderr(AsyncMode::Custom) => $layer
                    .with_writer(writer::async_writer_for(writer::AsyncWriterTarget::Stderr))
                    .boxed(),
                #[cfg(feature = "native-async")]
                WriterTarget::AsyncStderr(AsyncMode::Native) => $layer
                    .with_writer(writer::native_async_writer(
                        writer::AsyncWriterTarget::Stderr,
                    ))
                    .boxed(),
                #[cfg(feature = "file")]
                WriterTarget::File { .. } => $layer.with_writer(std::io::sink).boxed(),
            }
        }};
    }

    match &writer.format {
        Format::Pretty(cfg) => write_to!(
            base()
                .pretty()
                .with_target(cfg.target)
                .with_file(cfg.file)
                .with_line_number(cfg.line_number)
                .with_ansi(writer.ansi)
        ),
        Format::Compact(cfg) => write_to!(
            base()
                .with_target(cfg.target)
                .with_file(cfg.file)
                .with_line_number(cfg.line_number)
                .with_ansi(writer.ansi)
                .event_format(formatter)
        ),
        Format::Json(cfg) => write_to!(
            base()
                .json()
                .with_target(cfg.target)
                .with_file(cfg.file)
                .with_line_number(cfg.line_number)
                .with_current_span(cfg.current_span)
                .with_span_list(cfg.span_list)
                .flatten_event(cfg.flatten_event)
                .with_ansi(writer.ansi)
        ),
    }
}

pub fn build_layer<S>(writer: &Writer) -> FmtLayer<S>
where
    S: tracing::Subscriber + for<'a> tracing_subscriber::registry::LookupSpan<'a>,
{
    build_layer_inner(writer, None)
}

#[cfg(feature = "file")]
#[allow(clippy::single_call_fn)]
fn build_file_fmt_layer<S>(writer: &Writer, file_writer: writer::FileWriter) -> FmtLayer<S>
where
    S: tracing::Subscriber + for<'a> tracing_subscriber::registry::LookupSpan<'a>,
{
    let formatter = build_formatter(writer, ColorDepth::NoColor, None);
    let base = || {
        tracing_subscriber::fmt::Layer::default()
            .with_thread_ids(false)
            .with_thread_names(false)
            .with_span_events(FmtSpan::NONE)
    };

    match &writer.format {
        Format::Pretty(cfg) => base()
            .pretty()
            .with_target(cfg.target)
            .with_file(cfg.file)
            .with_line_number(cfg.line_number)
            .with_ansi(false)
            .with_writer(file_writer)
            .boxed(),
        Format::Compact(cfg) => base()
            .with_target(cfg.target)
            .with_file(cfg.file)
            .with_line_number(cfg.line_number)
            .with_ansi(false)
            .event_format(formatter)
            .with_writer(file_writer)
            .boxed(),
        Format::Json(cfg) => base()
            .json()
            .with_target(cfg.target)
            .with_file(cfg.file)
            .with_line_number(cfg.line_number)
            .with_current_span(cfg.current_span)
            .with_span_list(cfg.span_list)
            .flatten_event(cfg.flatten_event)
            .with_ansi(false)
            .with_writer(file_writer)
            .boxed(),
    }
}

/// Build a reload layer + TracingGuard.
///
/// `style_handle` can be:
/// - A `Style` (creates an independent handle — `with_style()` works on the
///   guard but won't affect `Formatter`s that already exist).
/// - An `Arc<ArcSwap<Style>>` from `Formatter::style_handle()` — the guard
///   and formatter share the same style state so `with_style()` updates the
///   formatter at runtime.
pub fn build_reload_filter(
    level: crate::config::Level,
    style_handle: impl Into<Arc<arc_swap::ArcSwap<Style>>>,
) -> (
    tracing_subscriber::reload::Layer<tracing_subscriber::EnvFilter, InnerSubscriber>,
    TracingGuard,
) {
    let filter = Filter::new(level);
    let env_filter =
        tracing_subscriber::EnvFilter::try_new(filter.as_directive()).unwrap_or_default();
    let (layer, raw) = ReloadHandle::new(env_filter);
    let guard = TracingGuard {
        raw,
        filter,
        style: style_handle.into(),
        #[cfg(feature = "file")]
        worker_guard: None,
        #[cfg(feature = "file")]
        log_path: None,
    };
    (layer, guard)
}

#[non_exhaustive]
pub struct SubscriberParts {
    pub(crate) raw: ReloadHandle,
    pub(crate) filter: Filter,
    pub(crate) style: Arc<arc_swap::ArcSwap<Style>>,
    #[cfg(feature = "file")]
    pub(crate) file_guard: Option<writer::LogHandle>,
    #[cfg(feature = "file")]
    pub(crate) log_path: Option<PathBuf>,
}

impl std::fmt::Debug for SubscriberParts {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("SubscriberParts").finish_non_exhaustive()
    }
}

impl SubscriberParts {
    pub fn into_guard(self) -> TracingGuard {
        TracingGuard {
            raw: self.raw,
            filter: self.filter,
            style: self.style,
            #[cfg(feature = "file")]
            worker_guard: self.file_guard,
            #[cfg(feature = "file")]
            log_path: self.log_path,
        }
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

#[allow(clippy::type_complexity)]
pub fn build_subscriber(config: Config) -> crate::Result<SubscriberParts> {
    let Config { level, writers } = config;
    let initial_style = writers.first().map_or_else(Style::default, |w| w.style);
    let shared_style = Arc::new(arc_swap::ArcSwap::new(Arc::new(initial_style)));

    let mut stdout_layer = None;
    let mut stderr_layer = None;
    #[cfg(any(feature = "custom-async", feature = "native-async"))]
    let mut async_stdout_layer = None;
    #[cfg(not(any(feature = "custom-async", feature = "native-async")))]
    let async_stdout_layer = None;
    #[cfg(any(feature = "custom-async", feature = "native-async"))]
    let mut async_stderr_layer = None;
    #[cfg(not(any(feature = "custom-async", feature = "native-async")))]
    let async_stderr_layer = None;
    #[cfg(feature = "file")]
    let mut file_layer = None;
    #[cfg(not(feature = "file"))]
    let file_layer = None;

    #[cfg(feature = "file")]
    let mut file_guard = None;
    #[cfg(feature = "file")]
    let mut log_path = None;

    let mut shared_style_used = false;

    for writer in writers {
        #[cfg(feature = "file")]
        let eligible = !matches!(writer.target, WriterTarget::File { .. });
        #[cfg(not(feature = "file"))]
        let eligible = true;

        let handle = (eligible && !shared_style_used).then(|| {
            shared_style_used = true;
            shared_style.clone()
        });

        match writer.target {
            WriterTarget::Stdout => stdout_layer = Some(build_layer_inner(&writer, handle)),
            WriterTarget::Stderr => stderr_layer = Some(build_layer_inner(&writer, handle)),
            #[cfg(feature = "file")]
            WriterTarget::File { ref path, rotation } => {
                let (file_w, guard, resolved_path) =
                    writer::file::build_file_layer(path, rotation)?;
                file_guard = Some(guard);
                log_path = Some(resolved_path);
                file_layer = Some(build_file_fmt_layer(&writer, file_w));
            }
            #[cfg(any(feature = "custom-async", feature = "native-async"))]
            WriterTarget::AsyncStdout(_) => {
                async_stdout_layer = Some(build_layer_inner(&writer, handle))
            }
            #[cfg(any(feature = "custom-async", feature = "native-async"))]
            WriterTarget::AsyncStderr(_) => {
                async_stderr_layer = Some(build_layer_inner(&writer, handle))
            }
        }
    }

    let filter = Filter::new(level);
    let env_filter =
        tracing_subscriber::EnvFilter::try_new(filter.as_directive()).unwrap_or_default();
    let (env_filter_layer, raw) = ReloadHandle::new(env_filter);

    let subscriber = tracing_subscriber::Registry::default()
        .with(stdout_layer)
        .with(stderr_layer)
        .with(async_stdout_layer)
        .with(async_stderr_layer)
        .with(file_layer)
        .with(env_filter_layer);

    let _ = tracing_log::LogTracer::init();
    tracing::subscriber::set_global_default(subscriber)?;
    Ok(SubscriberParts {
        raw,
        filter,
        style: shared_style,
        #[cfg(feature = "file")]
        file_guard,
        #[cfg(feature = "file")]
        log_path,
    })
}

pub fn init(config: impl Into<Config>) -> crate::Result<TracingGuard> {
    build_subscriber(config.into()).map(SubscriberParts::into_guard)
}

#[must_use = "dropping TracingGuard will release associated resources"]
#[allow(clippy::module_name_repetitions)]
#[non_exhaustive]
pub struct TracingGuard {
    pub(crate) raw: ReloadHandle,
    pub(crate) filter: Filter,
    pub(crate) style: Arc<arc_swap::ArcSwap<Style>>,
    #[cfg(feature = "file")]
    pub(crate) worker_guard: Option<writer::LogHandle>,
    #[cfg(feature = "file")]
    pub(crate) log_path: Option<PathBuf>,
}

impl std::fmt::Debug for TracingGuard {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let mut d = f.debug_struct("TracingGuard");
        d.field("filter", &self.filter);
        #[cfg(feature = "file")]
        d.field("log_path", &self.log_path);
        #[cfg(feature = "file")]
        d.field("has_file_guard", &self.worker_guard.is_some());
        d.finish_non_exhaustive()
    }
}

impl TracingGuard {
    pub fn with_style(&mut self, f: impl FnOnce(&mut Style)) {
        let mut style = **self.style.load();
        f(&mut style);
        self.style.store(Arc::new(style));
    }

    pub fn reload(&mut self, directive: &str) -> crate::Result<()> {
        self.apply_directive(directive)?;
        self.filter = Filter::new(crate::config::Level::Custom(directive.to_owned()));
        Ok(())
    }

    pub fn set_filter(&mut self, filter: Filter) -> crate::Result<()> {
        let directive = filter.as_directive();
        self.apply_directive(&directive)?;
        self.filter = filter;
        Ok(())
    }

    pub fn set_level(&mut self, level: impl Into<crate::config::Level>) -> crate::Result<()> {
        self.filter = Filter::new(level);
        self.apply_current_filter()
    }

    pub fn set_target_level(
        &mut self,
        target: impl Into<compact_str::CompactString>,
        level: impl Into<crate::config::Level>,
    ) -> crate::Result<()> {
        self.filter.with_target(target, level);
        self.apply_current_filter()
    }

    pub fn remove_target_level(&mut self, target: &str) -> crate::Result<()> {
        self.filter.remove_target(target);
        self.apply_current_filter()
    }

    fn apply_current_filter(&self) -> crate::Result<()> {
        let env_filter = tracing_subscriber::EnvFilter::try_new(self.filter.as_directive())?;
        self.raw.modify(|f| *f = env_filter)
    }

    fn apply_directive(&self, directive: &str) -> crate::Result<()> {
        let filter = tracing_subscriber::EnvFilter::try_new(directive)?;
        self.raw.modify(|f| *f = filter)
    }

    #[cfg(feature = "file")]
    pub fn log_path(&self) -> Option<&std::path::Path> {
        self.log_path.as_deref()
    }

    pub fn shutdown(&self) -> crate::Result<()> {
        let filter = tracing_subscriber::EnvFilter::try_new("off")?;
        self.raw.modify(|f| *f = filter)
    }
}
