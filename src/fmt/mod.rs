use chrono::Local;
use owo_colors::Style;
use std::fmt;
use std::io::Write;
use tracing::{Event, Level, Subscriber};
use tracing_subscriber::fmt::FormattedFields;
use tracing_subscriber::fmt::format::Writer;
use tracing_subscriber::fmt::{FmtContext, FormatEvent};
use tracing_subscriber::registry::LookupSpan;

use crate::config::LogRotation;
#[cfg(feature = "nerd")]
use nerd_font_symbols::{fa, oct, ple};

const BUILD_PATH_WIDTH: usize = match option_env!("SAGE_TRACE_MAX_PATH_WIDTH") {
    Some(s) => {
        let bytes = s.as_bytes();
        let mut i = 0;
        let mut n: usize = 0;
        while i < bytes.len() {
            n = n * 10 + (bytes[i] - b'0') as usize;
            i += 1;
        }
        n
    }
    None => 28,
};

#[derive(Clone, Debug)]
pub struct Icons {
    pub bracket_open: &'static str,
    pub bracket_close: &'static str,
    pub time_bracket_open: &'static str,
    pub time_bracket_close: &'static str,
    pub separator: &'static str,
    pub arrow: &'static str,
    pub span_delimiter: &'static str,
    pub span_join: &'static str,
}

impl Icons {
    pub const fn unicode() -> Self {
        Self {
            bracket_open: "[",
            bracket_close: "]",
            time_bracket_open: "「",
            time_bracket_close: "」",
            separator: "│",
            arrow: "❯",
            span_delimiter: "┇",
            span_join: "·",
        }
    }

    #[cfg(feature = "nerd")]
    pub const fn nerd() -> Self {
        Self {
            bracket_open: ple::PLE_LEFT_HALF_CIRCLE_THICK,
            bracket_close: ple::PLE_RIGHT_HALF_CIRCLE_THICK,
            time_bracket_open: "「",
            time_bracket_close: "」",
            separator: "\u{2502}",
            arrow: oct::OCT_CHEVRON_RIGHT,
            span_delimiter: fa::FA_CODE_MERGE,
            span_join: fa::FA_ANGLE_RIGHT,
        }
    }
}

impl Default for Icons {
    fn default() -> Self {
        #[cfg(feature = "nerd")]
        {
            Self::nerd()
        }
        #[cfg(all(feature = "unicode", not(feature = "nerd")))]
        {
            Self::unicode()
        }
        #[cfg(not(any(feature = "nerd", feature = "unicode")))]
        {
            Self::unicode()
        }
    }
}

#[derive(Clone, Debug)]
pub struct Theme {
    pub accent: Style,
    pub secondary: Style,
    pub text: Style,
    pub error: (u8, u8, u8),
    pub warn: (u8, u8, u8),
    pub info: (u8, u8, u8),
    pub debug: (u8, u8, u8),
    pub trace: (u8, u8, u8),
}

impl Theme {
    pub const fn trans_flag() -> Self {
        Self {
            accent: Style::new().truecolor(91, 206, 250),
            secondary: Style::new().truecolor(245, 169, 184),
            text: Style::new().truecolor(255, 255, 255),
            error: (255, 85, 85),
            warn: (255, 200, 60),
            info: (91, 206, 250),
            debug: (245, 169, 184),
            trace: (180, 180, 180),
        }
    }

    pub const fn monokai() -> Self {
        Self {
            accent: Style::new().truecolor(102, 217, 239),
            secondary: Style::new().truecolor(249, 38, 114),
            text: Style::new().truecolor(248, 248, 242),
            error: (255, 85, 85),
            warn: (255, 200, 60),
            info: (102, 217, 239),
            debug: (249, 38, 114),
            trace: (180, 180, 180),
        }
    }

    pub const fn dracula() -> Self {
        Self {
            accent: Style::new().truecolor(139, 233, 253),
            secondary: Style::new().truecolor(255, 121, 198),
            text: Style::new().truecolor(248, 248, 242),
            error: (255, 85, 85),
            warn: (255, 200, 60),
            info: (139, 233, 253),
            debug: (255, 121, 198),
            trace: (180, 180, 180),
        }
    }
    pub const fn nord() -> Self {
        Self {
            accent: Style::new().truecolor(136, 192, 208),
            secondary: Style::new().truecolor(163, 190, 140),
            text: Style::new().truecolor(216, 222, 233),
            error: (191, 97, 106),
            warn: (235, 203, 139),
            info: (136, 192, 208),
            debug: (163, 190, 140),
            trace: (180, 180, 180),
        }
    }

    pub const fn catppuccin_mocha() -> Self {
        Self {
            accent: Style::new().truecolor(137, 180, 250),
            secondary: Style::new().truecolor(203, 166, 247),
            text: Style::new().truecolor(205, 214, 244),
            error: (243, 139, 168),
            warn: (249, 226, 175),
            info: (137, 180, 250),
            debug: (203, 166, 247),
            trace: (180, 180, 180),
        }
    }

    pub const fn gruvbox() -> Self {
        Self {
            accent: Style::new().truecolor(131, 165, 152),
            secondary: Style::new().truecolor(254, 128, 25),
            text: Style::new().truecolor(235, 219, 178),
            error: (251, 73, 52),
            warn: (250, 189, 47),
            info: (131, 165, 152),
            debug: (254, 128, 25),
            trace: (180, 180, 180),
        }
    }

    pub const fn one_dark() -> Self {
        Self {
            accent: Style::new().truecolor(97, 175, 239),
            secondary: Style::new().truecolor(198, 120, 221),
            text: Style::new().truecolor(171, 178, 191),
            error: (224, 108, 117),
            warn: (229, 192, 123),
            info: (97, 175, 239),
            debug: (198, 120, 221),
            trace: (180, 180, 180),
        }
    }

    pub const fn tokyo_night() -> Self {
        Self {
            accent: Style::new().truecolor(122, 162, 247),
            secondary: Style::new().truecolor(187, 154, 247),
            text: Style::new().truecolor(192, 202, 245),
            error: (247, 118, 142),
            warn: (224, 175, 104),
            info: (122, 162, 247),
            debug: (187, 154, 247),
            trace: (180, 180, 180),
        }
    }
}

impl Default for Theme {
    fn default() -> Self {
        Self::trans_flag()
    }
}

#[derive(Clone, Debug)]
pub struct AnsiFormatter {
    pub time_format: String,
    pub path_width: usize,
    pub show_path: bool,
    pub show_spans: bool,
    pub theme: Theme,
    pub icons: Icons,
}

impl Default for AnsiFormatter {
    fn default() -> Self {
        Self {
            time_format: "%H:%M:%S".to_string(),
            path_width: BUILD_PATH_WIDTH,
            show_path: true,
            show_spans: true,
            theme: Theme::default(),
            icons: Icons::default(),
        }
    }
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
            return format!("{:>width$}", full, width = max_width);
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
            Level::ERROR => (theme.error, "ERROR"),
            Level::WARN => (theme.warn, "WARN "),
            Level::INFO => (theme.info, "INFO "),
            Level::DEBUG => (theme.debug, "DEBUG"),
            Level::TRACE => (theme.trace, "TRACE"),
        };

        // Darken color for text (divide by 2)
        let dark_rgb = (level_rgb.0 / 2, level_rgb.1 / 2, level_rgb.2 / 2);

        if is_nerd {
            write!(writer, "{}", theme.accent.style(icons.time_bracket_open))?;
            write!(writer, "{}", theme.text.style(time))?;
            write!(writer, "{} ", theme.accent.style(icons.time_bracket_close))?;
            write!(
                writer,
                "{}",
                Style::new()
                    .truecolor(level_rgb.0, level_rgb.1, level_rgb.2)
                    .style(icons.bracket_open)
            )?;
            write!(
                writer,
                "{}",
                Style::new()
                    .on_truecolor(level_rgb.0, level_rgb.1, level_rgb.2)
                    .truecolor(dark_rgb.0, dark_rgb.1, dark_rgb.2)
                    .bold()
                    .style(level_label)
            )?;
            write!(
                writer,
                "{} ",
                Style::new()
                    .truecolor(level_rgb.0, level_rgb.1, level_rgb.2)
                    .style(icons.bracket_close)
            )?;
        } else {
            write!(writer, "{}", theme.accent.style(icons.time_bracket_open))?;
            write!(writer, "{}", theme.text.style(time))?;
            write!(writer, "{} ", theme.accent.style(icons.time_bracket_close))?;
            write!(
                writer,
                "{} ",
                Style::new()
                    .on_truecolor(level_rgb.0, level_rgb.1, level_rgb.2)
                    .truecolor(dark_rgb.0, dark_rgb.1, dark_rgb.2)
                    .bold()
                    .style(level_label)
            )?;
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
        write!(writer, "{} ", theme.accent.dimmed().style(icons.separator))?;
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
            let trimmed = if val.starts_with('"') && val.ends_with('"') && val.len() >= 2 {
                val[1..val.len() - 1].to_string()
            } else {
                val
            };
            self.record_field(field.name(), trimmed);
        } else {
            self.record_field(field.name(), val);
        }
    }
}

#[cfg(test)]
mod test;

pub fn rotate_log_file(path: &std::path::Path, mode: LogRotation) -> anyhow::Result<()> {
    if !path.exists() {
        return Ok(());
    }

    match mode {
        LogRotation::None => Ok(()),
        LogRotation::Rename => {
            let timestamp = chrono::Local::now().format("%Y-%m-%d_%H-%M-%S");
            let renamed = path.with_extension(format!("{timestamp}.log"));
            std::fs::rename(path, renamed)?;
            Ok(())
        }
        LogRotation::Compress => {
            let timestamp = chrono::Local::now().format("%Y-%m-%d_%H-%M-%S");
            let gz_path = path.with_extension(format!("{timestamp}.log.gz"));
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
