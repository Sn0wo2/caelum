mod theme;

pub use theme::{Icons, LevelLabels, StyleConfig, Theme};

use chrono::Local;
use owo_colors::Style;
use std::borrow::Cow;
use std::fmt;
use std::sync::{Arc, PoisonError, RwLock};
use tracing::{Event, Level, Subscriber};
use tracing_subscriber::fmt::FormattedFields;
use tracing_subscriber::fmt::format::Writer;
use tracing_subscriber::fmt::{FmtContext, FormatEvent};
use tracing_subscriber::registry::LookupSpan;

const BUILD_PATH_WIDTH: usize = include!(concat!(env!("OUT_DIR"), "/path_width"));

#[derive(Clone, Debug, Default)]
pub struct AnsiFormatter {
    pub time_format: String,
    pub path_width: usize,
    pub show_path: bool,
    pub show_spans: bool,
    pub theme: Theme,
    pub icons: Icons,
    pub labels: LevelLabels,
    style: Arc<RwLock<StyleConfig>>,
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
            theme: config.theme,
            icons: config.icons,
            labels: config.labels,
            style: Arc::new(RwLock::new(config)),
        }
    }

    #[must_use]
    pub fn style_config(&self) -> Arc<RwLock<StyleConfig>> {
        self.style.clone()
    }

    #[must_use]
    pub fn with_theme(mut self, theme: Theme) -> Self {
        self.theme = theme;
        self.style
            .write()
            .unwrap_or_else(PoisonError::into_inner)
            .theme = theme;
        self
    }

    #[must_use]
    pub fn with_time_format(mut self, fmt: impl Into<String>) -> Self {
        self.time_format = fmt.into();
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

    #[must_use]
    pub fn with_icons(mut self, icons: Icons) -> Self {
        self.icons = icons;
        self.style
            .write()
            .unwrap_or_else(PoisonError::into_inner)
            .icons = icons;
        self
    }

    #[must_use]
    pub fn with_labels(mut self, labels: LevelLabels) -> Self {
        self.labels = labels;
        self.style
            .write()
            .unwrap_or_else(PoisonError::into_inner)
            .labels = labels;
        self
    }

    fn format_path(file: &str, line: u32, max_width: usize) -> String {
        let normalized: Cow<'_, str> = if file.contains('\\') {
            Cow::Owned(file.replace('\\', "/"))
        } else {
            Cow::Borrowed(file)
        };
        let stripped = normalized.find("src/").map_or(&*normalized, |i| {
            normalized
                .get(i.saturating_add(4)..)
                .unwrap_or(&normalized)
        });

        Self::smart_truncate(stripped, line, max_width)
    }

    fn smart_truncate(path: &str, line: u32, max_width: usize) -> String {
        let full = format!("{path}:{line}");
        if full.len() <= max_width {
            return format!("{full:<max_width$}");
        }

        if let Some(last_slash) = path.rfind('/') {
            let tail = path
                .get(last_slash.saturating_add(1)..)
                .unwrap_or(path);
            let file_part = format!("{tail}:{line}");
            if file_part.len().saturating_add(2) <= max_width {
                let dir_part = path.get(..last_slash).unwrap_or("");
                let remaining = max_width
                    .saturating_sub(file_part.len())
                    .saturating_sub(1);
                let dir_start = dir_part.len().saturating_sub(remaining);
                let dir_tail = dir_part.get(dir_start..).unwrap_or("");
                let clean_dir = dir_tail.find('/').map_or(dir_tail, |i| {
                    dir_tail.get(i.saturating_add(1)..).unwrap_or(dir_tail)
                });

                return format!("{clean_dir}/{file_part}");
            }
        }

        let start = full.len().saturating_sub(max_width.saturating_sub(1));
        format!("…{}", full.get(start..).unwrap_or(""))
    }

    fn resolve_style(&self) -> (Theme, Icons, LevelLabels) {
        let style = self.style.read().unwrap_or_else(PoisonError::into_inner);
        (style.theme, style.icons, style.labels)
    }

    fn format_path_section(
        &self,
        writer: &mut Writer<'_>,
        event: &Event<'_>,
        theme: &Theme,
        icons: &Icons,
    ) -> fmt::Result {
        let file = event.metadata().file().unwrap_or("?");
        let line = event.metadata().line().unwrap_or(0);
        write!(
            writer,
            "{}",
            theme
                .text
                .dimmed()
                .style(Self::format_path(file, line, self.path_width))
        )?;
        write!(writer, " {} ", theme.accent.style(icons.arrow))?;
        Ok(())
    }

    fn format_fields(
        writer: &mut Writer<'_>,
        event: &Event<'_>,
        theme: &Theme,
    ) -> fmt::Result {
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
                    write!(writer, "{}", theme.accent.style(icons.span_join))?;
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
        N: for<'b> tracing_subscriber::fmt::FormatFields<'b> + 'static,
    {
        write!(writer, "{}", theme.text.dimmed().style(name))?;

        if let Some(fields) = fields {
            let fields = fields.fields.as_str();
            if !fields.is_empty() {
                write!(writer, "{}", theme.accent.dimmed().style("{"))?;
                write!(writer, "{}", theme.text.dimmed().style(fields))?;
                write!(writer, "{}", theme.accent.dimmed().style("}"))?;
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
        let (theme, icons, labels) = self.resolve_style();
        let is_nerd = icons.is_nerd();

        let level = event.metadata().level();

        let time = Local::now().format(&self.time_format).to_string();
        let (level_rgb, level_label) = match *level {
            Level::ERROR => (theme.error, labels.error),
            Level::WARN => (theme.warn, labels.warn),
            Level::INFO => (theme.info, labels.info),
            Level::DEBUG => (theme.debug, labels.debug),
            Level::TRACE => (theme.trace, labels.trace),
        };

        // RGB darkening — halve each channel via bit shift (equivalent to /2, avoids integer_division lint).
        let dark_rgb = (level_rgb.0 >> 1, level_rgb.1 >> 1, level_rgb.2 >> 1);
        let fg_style = Style::new().truecolor(level_rgb.0, level_rgb.1, level_rgb.2);
        let bg_style = Style::new()
            .on_truecolor(level_rgb.0, level_rgb.1, level_rgb.2)
            .truecolor(dark_rgb.0, dark_rgb.1, dark_rgb.2)
            .bold();

        write!(writer, "{}", theme.accent.style(icons.time_bracket_open))?;
        write!(writer, "{}", theme.text.style(time))?;

        if is_nerd {
            write!(writer, " {} ", theme.accent.dimmed().style(icons.separator))?;
            write!(writer, "{}", fg_style.style(icons.bracket_open))?;
            write!(writer, "{}", bg_style.style(level_label))?;
            write!(writer, "{} ", fg_style.style(icons.bracket_close))?;
            write!(writer, "{} ", theme.accent.style(icons.time_bracket_close))?;
        } else {
            write!(writer, "{} ", theme.accent.style(icons.time_bracket_close))?;
            write!(writer, "{} ", bg_style.style(level_label))?;
        }

        if self.show_path {
            self.format_path_section(&mut writer, event, &theme, &icons)?;
        }

        Self::format_fields(&mut writer, event, &theme)?;

        if self.show_spans {
            Self::format_spans(&mut writer, ctx, &theme, &icons)?;
        }

        writeln!(writer)
    }
}

#[derive(Default)]
struct EventVisitor {
    message: Option<String>,
    fields: Vec<(String, String)>,
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
    fn record_str(&mut self, field: &tracing::field::Field, value: &str) {
        self.record_field(field.name(), value.to_owned());
    }

    fn record_debug(&mut self, field: &tracing::field::Field, value: &dyn fmt::Debug) {
        let val = format!("{value:?}");
        let trimmed = if matches!(field.name(), "message" | "msg") {
            val.strip_prefix('"')
                .and_then(|s| s.strip_suffix('"'))
                .map(str::to_owned)
                .unwrap_or(val)
        } else {
            val
        };
        self.record_field(field.name(), trimmed);
    }
}

#[cfg(test)]
mod test;
