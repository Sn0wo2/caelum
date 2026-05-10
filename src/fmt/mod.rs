mod theme;

pub use theme::{Icons, LevelLabels, StyleConfig, Theme};

use arrayvec::ArrayString;
use chrono::Local;
use owo_colors::Style;
use smallvec::SmallVec;
use std::borrow::Cow;
use std::fmt;
use std::fmt::{Debug, Write};
use std::sync::Mutex;

use tracing::field::Field;
use tracing::{Event, Level, Subscriber};
use tracing_subscriber::fmt::FormattedFields;
use tracing_subscriber::fmt::format::Writer;
use tracing_subscriber::fmt::{FmtContext, FormatEvent};
use tracing_subscriber::registry::LookupSpan;

const BUILD_PATH_WIDTH: usize = include!(concat!(env!("OUT_DIR"), "/path_width"));

const PATH_BUF_SIZE: usize = 256;

#[derive(Debug)]
pub struct AnsiFormatter {
    pub time_format: String,
    pub path_width: usize,
    pub show_path: bool,
    pub show_spans: bool,
    style: StyleConfig,
    #[doc(hidden)]
    cached_time: Mutex<(i64, String)>,
}

impl Clone for AnsiFormatter {
    fn clone(&self) -> Self {
        Self {
            time_format: self.time_format.clone(),
            path_width: self.path_width,
            show_path: self.show_path,
            show_spans: self.show_spans,
            style: self.style,
            cached_time: Mutex::new(self.cached_time.lock().unwrap().clone()),
        }
    }
}

impl Default for AnsiFormatter {
    fn default() -> Self {
        Self::new()
    }
}

impl AnsiFormatter {
    #[must_use]
    pub fn new() -> Self {
        let config = StyleConfig::default();
        Self {
            time_format: String::from("%H:%M:%S"),
            path_width: BUILD_PATH_WIDTH,
            show_path: true,
            show_spans: true,
            style: config,
            cached_time: Mutex::new((0, String::new())),
        }
    }

    #[must_use]
    pub fn style_config(&self) -> &StyleConfig {
        &self.style
    }

    #[must_use]
    pub fn with_style_config(mut self, style: StyleConfig) -> Self {
        self.style = style;
        self
    }

    #[must_use]
    pub fn with_icons(mut self, icons: Icons) -> Self {
        self.style.icons = icons;
        self
    }

    #[must_use]
    pub fn with_labels(mut self, labels: LevelLabels) -> Self {
        self.style.labels = labels;
        self
    }

    #[must_use]
    pub fn with_theme(mut self, theme: Theme) -> Self {
        self.style.theme = theme;
        self
    }

    #[must_use]
    pub fn with_time_format(mut self, fmt: impl Into<String>) -> Self {
        self.time_format = fmt.into();
       *self.cached_time.lock().unwrap() = (0, String::new());
        self
    }

    #[must_use]
    pub const fn with_path_width(mut self, width: usize) -> Self {
        self.path_width = width;
        self
    }

    #[must_use]
    pub const fn with_show_path(mut self, show: bool) -> Self {
        self.show_path = show;
        self
    }

    #[must_use]
    pub const fn with_show_spans(mut self, show: bool) -> Self {
        self.show_spans = show;
        self
    }

   fn get_time(&self) -> String {
        let now = Local::now();
        let ts = now.timestamp_millis();
        let mut cache = self.cached_time.lock().unwrap();
        if ts != cache.0 {
            *cache = (ts, now.format(&self.time_format).to_string());
        }
        cache.1.clone()
    }

    fn format_path(file: &str, line: u32, max_width: usize) -> ArrayString<PATH_BUF_SIZE> {
        let normalized: Cow<'_, str> = if file.contains('\\') {
            Cow::Owned(file.replace('\\', "/"))
        } else {
            Cow::Borrowed(file)
        };
        Self::smart_truncate(
            normalized
                .find("src/")
                .map_or(&*normalized, |i| {
                    normalized.get(i.saturating_add(4)..).unwrap_or(&normalized)
                }),
            line,
            max_width,
        )
    }

    fn smart_truncate(
        path: &str,
        line: u32,
        max_width: usize,
    ) -> ArrayString<PATH_BUF_SIZE> {
        let mut full = ArrayString::<PATH_BUF_SIZE>::new();
        let _ = write!(full, "{path}:{line}");

        if full.len() <= max_width {
            let mut result = ArrayString::<PATH_BUF_SIZE>::new();
            let pad = max_width - full.len();
            for _ in 0..pad {
                result.push(' ');
            }
            result.push_str(full.as_str());
            return result;
        }

        if let Some(last_slash) = path.rfind('/') {
            let tail = path.get(last_slash.saturating_add(1)..).unwrap_or(path);
            let mut file_part = ArrayString::<PATH_BUF_SIZE>::new();
            let _ = write!(file_part, "{tail}:{line}");

            if file_part.len().saturating_add(2) <= max_width {
                let dir_part = path.get(..last_slash).unwrap_or("");
                let dir_start = dir_part.len().saturating_sub(
                    max_width.saturating_sub(file_part.len()).saturating_sub(1),
                );
                let dir_tail = dir_part.get(dir_start..).unwrap_or("");
                let clean_dir = dir_tail.find('/').map_or(dir_tail, |i| {
                    dir_tail.get(i.saturating_add(1)..).unwrap_or(dir_tail)
                });

                let mut result = ArrayString::<PATH_BUF_SIZE>::new();
                let _ = write!(result, "{clean_dir}/{file_part}");
                return result;
            }
        }

        let start = full.len().saturating_sub(max_width.saturating_sub(1));
        let mut result = ArrayString::<PATH_BUF_SIZE>::new();
        result.push('\u{2026}'); // … ellipsis
        result.push_str(full.get(start..).unwrap_or(""));
        result
    }

    fn format_path_section(
        &self,
        writer: &mut Writer<'_>,
        event: &Event<'_>,
        theme: &Theme,
        icons: &Icons,
    ) -> fmt::Result {
        write!(
            writer,
            "{}",
            theme.text_dimmed.style(Self::format_path(
                event.metadata().file().unwrap_or("?"),
                event.metadata().line().unwrap_or(0),
                self.path_width
            ))
        )?;
        write!(writer, " {} ", theme.accent.style(icons.arrow))?;
        Ok(())
    }

    fn format_fields(writer: &mut Writer<'_>, event: &Event<'_>, theme: &Theme) -> fmt::Result {
        let mut visitor = EventVisitor::default();
        event.record(&mut visitor);

        let mut need_space = false;
        if let Some(msg) = visitor.message {
            write!(writer, "{}", theme.text.style(msg))?;
            need_space = !visitor.fields.is_empty();
        }

        for (k, v) in visitor.fields {
            if need_space {
                write!(writer, " ")?;
            }
            write!(
                writer,
                "{}{}{}",
                theme.secondary.style(k),
                theme.accent.style("="),
                theme.text.style(v)
            )?;
            need_space = true;
        }
        Ok(())
    }

    fn format_spans<S, N>(
        writer: &mut Writer<'_>,
        ctx: &FmtContext<'_, S, N>,
        theme: &Theme,
        icons: &Icons,
    ) -> fmt::Result
    where
        S: Subscriber + for<'a> LookupSpan<'a>,
        N: for<'a> tracing_subscriber::fmt::FormatFields<'a> + 'static,
    {
        if let Some(scope) = ctx.event_scope() {
            write!(writer, " {} ", theme.accent.style(icons.span_delimiter))?;
            for (i, span) in scope.from_root().enumerate() {
                if i > 0 {
                    write!(writer, " {} ", theme.accent.style(icons.span_join))?;
                }
                Self::write_span(
                    writer.by_ref(),
                    span.name(),
                    span.extensions().get::<FormattedFields<N>>(),
                    theme,
                )?;
            }
        } else if let Some(span) = ctx.lookup_current() {
            write!(writer, " {} ", theme.accent.style(icons.span_delimiter))?;
            Self::write_span(
                writer.by_ref(),
                span.name(),
                span.extensions().get::<FormattedFields<N>>(),
                theme,
            )?;
        }
        Ok(())
    }

    fn write_span<N>(
        mut writer: Writer<'_>,
        name: &str,
        fields: Option<&FormattedFields<N>>,
        theme: &Theme,
    ) -> fmt::Result
    where
        N: for<'a> tracing_subscriber::fmt::FormatFields<'a> + 'static,
    {
        write!(writer, "{}", theme.text_dimmed.style(name))?;

        if let Some(fields) = fields {
            let fields = fields.fields.as_str();
            if !fields.is_empty() {
                write!(writer, "{}", theme.accent_dimmed.style("{"))?;
                write!(writer, "{}", theme.text_dimmed.style(fields))?;
                write!(writer, "{}", theme.accent_dimmed.style("}"))?;
            }
        }
        Ok(())
    }
}

impl<S, N> FormatEvent<S, N> for AnsiFormatter
where
    S: Subscriber + for<'a> LookupSpan<'a>,
    N: for<'a> tracing_subscriber::fmt::FormatFields<'a> + 'static,
{
    fn format_event(
        &self,
        ctx: &FmtContext<'_, S, N>,
        mut writer: Writer<'_>,
        event: &Event<'_>,
    ) -> fmt::Result {
        let config = self.style_config();
        let theme = &config.theme;
        let icons = &config.icons;
        let labels = &config.labels;
        let is_nerd = icons.is_nerd();

        let level = event.metadata().level();

        let time = self.get_time();

        let lc = match *level {
            Level::ERROR => &theme.error,
            Level::WARN => &theme.warn,
            Level::INFO => &theme.info,
            Level::DEBUG => &theme.debug,
            Level::TRACE => &theme.trace,
        };
        let level_label = match *level {
            Level::ERROR => labels.error,
            Level::WARN => labels.warn,
            Level::INFO => labels.info,
            Level::DEBUG => labels.debug,
            Level::TRACE => labels.trace,
        };

        let mut fg_style = Style::new().truecolor(lc.rgb.0, lc.rgb.1, lc.rgb.2);
        if !is_nerd {
            fg_style = fg_style.on_truecolor(lc.rgb.0, lc.rgb.1, lc.rgb.2);
        }
        let bg_style = lc.bg; // pre-computed — no construction on hot path

         write!(writer, "{}", theme.accent.style(icons.time_bracket_open))?;
        write!(writer, "{}", theme.text.style(time))?;
        write!(
            writer,
            " {} ",
            theme.accent_dimmed.style(icons.separator)
        )?;

        write!(writer, "{}", fg_style.style(icons.bracket_open))?;
        write!(writer, "{}", bg_style.style(level_label))?;
        write!(writer, "{} ", fg_style.style(icons.bracket_close))?;

        write!(
            writer,
            "{} ",
            theme.accent.style(icons.time_bracket_close)
        )?;

        if self.show_path {
            self.format_path_section(&mut writer, event, theme, icons)?;
        }

        Self::format_fields(&mut writer, event, theme)?;

        if self.show_spans {
            Self::format_spans(&mut writer, ctx, theme, icons)?;
        }

        writeln!(writer)
    }
}
#[derive(Default)]
struct EventVisitor {
    message: Option<String>,
    fields: SmallVec<[(String, String); 4]>,
}

impl EventVisitor {
    fn record_field(&mut self, name: &str, value: String) {
        if name == "message" || name == "msg" {
            self.message = Some(value);
        } else {
            self.fields.push((name.to_owned(), value));
        }
    }
}

impl tracing::field::Visit for EventVisitor {
    fn record_str(&mut self, field: &Field, value: &str) {
        self.record_field(field.name(), value.to_owned());
    }

    fn record_debug(&mut self, field: &Field, value: &dyn Debug) {
        self.record_field(field.name(), format!("{:?}", value));
    }
}

#[cfg(test)]
mod test;
