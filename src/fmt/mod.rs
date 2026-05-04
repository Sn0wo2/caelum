mod theme;

pub use theme::{Icons, LevelLabels, Theme};

use chrono::Local;
use owo_colors::Style;
use smart_default::SmartDefault;
use std::fmt;
use tracing::{Event, Level, Subscriber};
use tracing_subscriber::fmt::FormattedFields;
use tracing_subscriber::fmt::format::Writer;
use tracing_subscriber::fmt::{FmtContext, FormatEvent};
use tracing_subscriber::registry::LookupSpan;

const BUILD_PATH_WIDTH: usize = include!(concat!(env!("OUT_DIR"), "/path_width"));

#[derive(Clone, Debug, SmartDefault)]
pub struct AnsiFormatter {
    #[default("%H:%M:%S".to_string())]
    pub time_format: String,
    #[default(BUILD_PATH_WIDTH)]
    pub path_width: usize,
    #[default = true]
    pub show_path: bool,
    #[default = true]
    pub show_spans: bool,
    #[default(Theme::default())]
    pub theme: Theme,
    #[default(Icons::default())]
    pub icons: Icons,
    #[default(LevelLabels::default())]
    pub labels: LevelLabels,
}

impl AnsiFormatter {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_theme(mut self, theme: Theme) -> Self {
        self.theme = theme;
        self
    }

    pub fn with_time_format(mut self, fmt: impl Into<String>) -> Self {
        self.time_format = fmt.into();
        self
    }

    pub fn with_path_width(mut self, width: usize) -> Self {
        self.path_width = width;
        self
    }

    pub fn with_show_path(mut self, show: bool) -> Self {
        self.show_path = show;
        self
    }

    pub fn with_show_spans(mut self, show: bool) -> Self {
        self.show_spans = show;
        self
    }

    pub fn with_icons(mut self, icons: Icons) -> Self {
        self.icons = icons;
        self
    }

    pub fn with_labels(mut self, labels: LevelLabels) -> Self {
        self.labels = labels;
        self
    }

    fn format_path(file: &str, line: u32, max_width: usize) -> String {
        let normalized = file.replace('\\', "/");
        let stripped = normalized
            .find("src/")
            .map(|i| &normalized[i + 4..])
            .unwrap_or(&normalized);

        Self::smart_truncate(stripped, line, max_width)
    }

    fn smart_truncate(path: &str, line: u32, max_width: usize) -> String {
        let full = format!("{}:{}", path, line);
        if full.len() <= max_width {
            return format!("{:<width$}", full, width = max_width);
        }

        if let Some(last_slash) = path.rfind('/') {
            let file_part = format!("{}:{}", &path[last_slash + 1..], line);
            if file_part.len() + 2 <= max_width {
                let dir_part = &path[..last_slash];
                let remaining = max_width - file_part.len() - 1;
                let dir_start = dir_part.len().saturating_sub(remaining);
                let clean_dir = dir_part[dir_start..]
                    .find('/')
                    .map(|i| &dir_part[dir_start + i + 1..])
                    .unwrap_or(&dir_part[dir_start..]);

                return format!("{}/{}", clean_dir, file_part);
            }
        }

        format!("…{}", &full[full.len().saturating_sub(max_width - 1)..])
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
        let theme = &self.theme;
        let icons = &self.icons;
        let is_nerd = icons.bracket_open != "[";

        let level = event.metadata().level();

        let time = Local::now().format(&self.time_format).to_string();
        let (level_rgb, level_label) = match *level {
            Level::ERROR => (theme.error, self.labels.error),
            Level::WARN => (theme.warn, self.labels.warn),
            Level::INFO => (theme.info, self.labels.info),
            Level::DEBUG => (theme.debug, self.labels.debug),
            Level::TRACE => (theme.trace, self.labels.trace),
        };

        let dark_rgb = (level_rgb.0 / 2, level_rgb.1 / 2, level_rgb.2 / 2);
        let level_fg = Style::new().truecolor(level_rgb.0, level_rgb.1, level_rgb.2);
        let level_bg = Style::new()
            .on_truecolor(level_rgb.0, level_rgb.1, level_rgb.2)
            .truecolor(dark_rgb.0, dark_rgb.1, dark_rgb.2)
            .bold();

        write!(writer, "{}", theme.accent.style(icons.time_bracket_open))?;
        write!(writer, "{}", theme.text.style(time))?;

        if is_nerd {
            write!(writer, " {} ", theme.accent.dimmed().style(icons.separator))?;
            write!(writer, "{}", level_fg.style(icons.bracket_open))?;
            write!(writer, "{}", level_bg.style(level_label))?;
            write!(writer, "{} ", level_fg.style(icons.bracket_close))?;
            write!(writer, "{} ", theme.accent.style(icons.time_bracket_close))?;
        } else {
            write!(writer, "{} ", theme.accent.style(icons.time_bracket_close))?;
            write!(writer, "{} ", level_bg.style(level_label))?;
        }

        if self.show_path {
            self.format_path_section(&mut writer, event, theme, icons)?;
        }

        self.format_fields(&mut writer, event, theme)?;

        if self.show_spans {
            self.format_spans(&mut writer, ctx, theme, icons)?;
        }

        writeln!(writer)
    }
}

impl AnsiFormatter {
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
        &self,
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
        &self,
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
                self.write_span(
                    writer.by_ref(),
                    span.name(),
                    span.extensions().get::<FormattedFields<N>>(),
                )?;
            }
        } else if let Some(span) = ctx.lookup_current() {
            write!(writer, " {} ", theme.accent.style(icons.span_delimiter))?;
            self.write_span(
                writer.by_ref(),
                span.name(),
                span.extensions().get::<FormattedFields<N>>(),
            )?;
        }
        Ok(())
    }

    fn write_span<N>(
        &self,
        mut writer: Writer<'_>,
        name: &str,
        fields: Option<&FormattedFields<N>>,
    ) -> fmt::Result
    where
        N: for<'b> tracing_subscriber::fmt::FormatFields<'b> + 'static,
    {
        let theme = &self.theme;
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
            self.fields.push((name.to_string(), value));
        }
    }
}

impl tracing::field::Visit for EventVisitor {
    fn record_str(&mut self, field: &tracing::field::Field, value: &str) {
        self.record_field(field.name(), value.to_string());
    }

    fn record_debug(&mut self, field: &tracing::field::Field, value: &dyn fmt::Debug) {
        let val = format!("{:?}", value);
        if field.name() == "message" || field.name() == "msg" {
            let trimmed = val
                .strip_prefix('"')
                .and_then(|s| s.strip_suffix('"'))
                .map(str::to_string)
                .unwrap_or(val);
            self.record_field(field.name(), trimmed);
        } else {
            self.record_field(field.name(), val);
        }
    }
}

#[cfg(test)]
mod test;
