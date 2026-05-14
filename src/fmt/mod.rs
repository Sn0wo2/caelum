use crate::config::{Icons, LevelLabels, StyleConfig, Theme};
use arrayvec::ArrayString;
use chrono::Utc;
use owo_colors::Style;
use smallvec::SmallVec;
use std::borrow::Cow;
use std::fmt;
use std::fmt::{Debug, Write};

use tracing::field::Field;
use tracing::{Event, Level, Subscriber};
use tracing_subscriber::fmt::FormattedFields;
use tracing_subscriber::fmt::format::Writer;
use tracing_subscriber::fmt::{FmtContext, FormatEvent};
use tracing_subscriber::registry::LookupSpan;

#[derive(Clone, Copy, Debug)]
struct LevelStyles {
    bracket_fg: Style,
    bracket_bg: Style,
    label: Style,
}

#[allow(clippy::missing_const_for_fn)]
fn make_level_styles(r: u8, g: u8, b: u8) -> LevelStyles {
    LevelStyles {
        bracket_fg: Style::new().truecolor(r, g, b),
        bracket_bg: Style::new().truecolor(r, g, b).on_truecolor(r, g, b),
        label: Style::new()
            .truecolor(r >> 2, g >> 2, b >> 2)
            .on_truecolor(r, g, b),
    }
}

fn build_all_level_styles(theme: &Theme) -> [LevelStyles; 5] {
    [
        make_level_styles(theme.error.0, theme.error.1, theme.error.2),
        make_level_styles(theme.warn.0, theme.warn.1, theme.warn.2),
        make_level_styles(theme.info.0, theme.info.1, theme.info.2),
        make_level_styles(theme.debug.0, theme.debug.1, theme.debug.2),
        make_level_styles(theme.trace.0, theme.trace.1, theme.trace.2),
    ]
}

const BUILD_PATH_WIDTH: usize = include!(concat!(env!("OUT_DIR"), "/path_width"));

const PATH_BUF_SIZE: usize = 256;

#[derive(Debug, Clone)]
pub struct AnsiFormatter {
    pub(crate) time_format: String,
    pub(crate) path_width: usize,
    pub(crate) show_path: bool,
    pub(crate) show_spans: bool,
    style: StyleConfig,
    level_styles: [LevelStyles; 5],
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
        let level_styles = build_all_level_styles(&config.theme);
        Self {
            time_format: String::from("%H:%M:%S"),
            path_width: BUILD_PATH_WIDTH,
            show_path: true,
            show_spans: true,
            style: config,
            level_styles,
        }
    }

    #[must_use]
    pub const fn style_config(&self) -> &StyleConfig {
        &self.style
    }

    #[must_use]
    pub const fn style_config_mut(&mut self) -> &mut StyleConfig {
        &mut self.style
    }

    #[must_use]
    pub fn with_style_config(mut self, style: StyleConfig) -> Self {
        self.level_styles = build_all_level_styles(&style.theme);
        self.style = style;
        self
    }

    #[must_use]
    pub const fn with_icons(mut self, icons: Icons) -> Self {
        self.style.icons = icons;
        self
    }

    #[must_use]
    pub const fn with_labels(mut self, labels: LevelLabels) -> Self {
        self.style.labels = labels;
        self
    }

    #[must_use]
    pub fn with_theme(mut self, theme: Theme) -> Self {
        self.level_styles = build_all_level_styles(&theme);
        self.style.theme = theme;
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

    fn write_time(&self, writer: &mut Writer<'_>, theme: &Theme) -> fmt::Result {
        let now = Utc::now();
        write!(
            writer,
            "{}",
            theme.text.style(now.format(&self.time_format))
        )
    }
    #[allow(clippy::single_call_fn)]
    fn format_path(file: &str, line: u32, max_width: usize) -> ArrayString<PATH_BUF_SIZE> {
        let normalized: Cow<'_, str> = if file.contains('\\') {
            Cow::Owned(file.replace('\\', "/"))
        } else {
            Cow::Borrowed(file)
        };
        Self::smart_truncate(
            normalized.find("src/").map_or(&*normalized, |i| {
                normalized.get(i.saturating_add(4)..).unwrap_or(&normalized)
            }),
            line,
            max_width,
        )
    }
    #[allow(clippy::single_call_fn)]
    fn smart_truncate(path: &str, line: u32, max_width: usize) -> ArrayString<PATH_BUF_SIZE> {
        let mut full = ArrayString::<PATH_BUF_SIZE>::new();
        let _ = write!(full, "{path}:{line}");

        if full.len() <= max_width {
            let mut result = ArrayString::<PATH_BUF_SIZE>::new();
            write!(result, "{full:>max_width$}").ok();
            return result;
        }

        if let Some(last_slash) = path.rfind('/') {
            let tail = path.get(last_slash.saturating_add(1)..).unwrap_or(path);
            let mut file_part = ArrayString::<PATH_BUF_SIZE>::new();
            let _ = write!(file_part, "{tail}:{line}");

            if file_part.len().saturating_add(2) <= max_width {
                let dir_part = path.get(..last_slash).unwrap_or("");
                let dir_start = dir_part
                    .len()
                    .saturating_sub(max_width.saturating_sub(file_part.len()).saturating_sub(1));
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
        result.push('\u{2026}');
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
            theme.text.dimmed().style(Self::format_path(
                event.metadata().file().unwrap_or("?"),
                event.metadata().line().unwrap_or(0),
                self.path_width
            ))
        )?;
        write!(writer, " {} ", theme.accent.style(icons.arrow))?;
        Ok(())
    }
    #[allow(clippy::single_call_fn)]
    fn format_fields(writer: &mut Writer<'_>, event: &Event<'_>, theme: &Theme) -> fmt::Result {
        let mut visitor = EventVisitor::default();
        event.record(&mut visitor);

        let mut sep = if let Some(msg) = visitor.message {
            write!(writer, "{}", theme.text.style(msg))?;
            " "
        } else {
            ""
        };

        for (k, v) in visitor.fields {
            write!(
                writer,
                "{}{}{}{}",
                sep,
                theme.secondary.style(k),
                theme.accent.style("="),
                theme.text.style(v)
            )?;
            sep = " ";
        }

        Ok(())
    }
    #[allow(clippy::single_call_fn)]
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
        let Some(scope) = ctx
            .event_scope()
            .or_else(|| ctx.lookup_current().map(|s| s.scope()))
        else {
            return Ok(());
        };

        let spans: SmallVec<[_; 8]> = scope.from_root().collect();
        if spans.is_empty() {
            return Ok(());
        }

        let total = spans.len();
        let accent = theme.accent;
        let text = theme.text;

        write!(writer, " {}", accent.style("["))?;

        for (i, span) in spans.iter().enumerate() {
            if i > 0 {
                write!(writer, "{} ", accent.dimmed().style(icons.span_join))?;
            }

            let span_style = if i == total - 1 { text } else { text.dimmed() };

            write!(writer, "{}", span_style.style(span.name()))?;

            let extensions = span.extensions();
            if let Some(fields) = extensions.get::<FormattedFields<N>>() {
                let fields_str = fields.fields.as_str();
                if !fields_str.is_empty() {
                    write!(writer, " {}", span_style.style(fields_str))?;
                }
            }
        }

        write!(writer, "{}", accent.style("]"))
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
        let is_nerd = config.icons.is_nerd();

        let level = event.metadata().level();

        let (ls, level_label) = match *level {
            Level::ERROR => (&self.level_styles[0], config.labels.error),
            Level::WARN => (&self.level_styles[1], config.labels.warn),
            Level::INFO => (&self.level_styles[2], config.labels.info),
            Level::DEBUG => (&self.level_styles[3], config.labels.debug),
            Level::TRACE => (&self.level_styles[4], config.labels.trace),
        };

        let fg_style = if is_nerd {
            ls.bracket_fg
        } else {
            ls.bracket_bg
        };

        write!(
            writer,
            "{}",
            config.theme.accent.style(config.icons.time_bracket_open)
        )?;
        self.write_time(&mut writer, &config.theme)?;
        write!(
            writer,
            " {} ",
            config.theme.accent.dimmed().style(config.icons.separator)
        )?;

        write!(writer, "{}", fg_style.style(config.icons.bracket_open))?;
        write!(writer, "{}", ls.label.style(level_label))?;
        write!(writer, "{} ", fg_style.style(config.icons.bracket_close))?;

        write!(
            writer,
            "{} ",
            config.theme.accent.style(config.icons.time_bracket_close)
        )?;

        if self.show_path {
            self.format_path_section(&mut writer, event, &config.theme, &config.icons)?;
        }

        Self::format_fields(&mut writer, event, &config.theme)?;

        if self.show_spans {
            Self::format_spans(&mut writer, ctx, &config.theme, &config.icons)?;
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
        self.record_field(field.name(), format!("{value:?}"));
    }
}

#[cfg(test)]
mod test;
