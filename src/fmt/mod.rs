use crate::color::style::{rgb_to_owo, rgb_to_owo_on, theme_fg_dimmed};
use crate::config::ColorDepth;
use crate::config::{Icons, LevelLabels, Style, Theme};
use arrayvec::ArrayString;
use chrono::Utc;
use owo_colors::Style as OwoStyle;
use smallvec::SmallVec;

use std::fmt;
use std::fmt::Write;
use std::sync::Arc;

use tracing::{Event, Level, Subscriber};
use tracing_subscriber::fmt::FormattedFields;
use tracing_subscriber::fmt::format::Writer;
use tracing_subscriber::fmt::{FmtContext, FormatEvent};
use tracing_subscriber::registry::LookupSpan;

mod visitor;
use visitor::EventVisitor;

#[derive(Clone, Copy, Debug)]
pub(crate) struct LevelStyles {
    bracket_fg: OwoStyle,
    bracket_bg: OwoStyle,
    label: OwoStyle,
}

const BUILD_PATH_WIDTH: usize = include!(concat!(env!("OUT_DIR"), "/path_width"));

const PATH_BUF_SIZE: usize = 256;

#[derive(Debug, Clone)]
#[non_exhaustive]
pub struct Formatter {
    pub(crate) time_format: String,
    pub(crate) path_width: usize,
    pub(crate) show_path: bool,
    pub(crate) show_spans: bool,
    pub(crate) style: Arc<arc_swap::ArcSwap<Style>>,
    pub(crate) color_depth: ColorDepth,
}

impl Default for Formatter {
    fn default() -> Self {
        Self::new()
    }
}

impl Formatter {
    #[must_use]
    pub fn new() -> Self {
        Self {
            time_format: String::from("%H:%M:%S"),
            path_width: BUILD_PATH_WIDTH,
            show_path: true,
            show_spans: true,
            style: Arc::new(arc_swap::ArcSwap::new(Arc::new(Style::default()))),
            color_depth: ColorDepth::TrueColor,
        }
    }

    /// Creates a Formatter that shares its style state with an existing handle.
    /// All builder methods (`with_theme`, `with_icons`, etc.) will write through
    /// this handle, and any future writes to the handle (e.g. via TracingGuard::with_style)
    /// are visible immediately in `format_event`.
    #[must_use]
    pub fn new_with_handle(style: Arc<arc_swap::ArcSwap<Style>>) -> Self {
        Self {
            time_format: String::from("%H:%M:%S"),
            path_width: BUILD_PATH_WIDTH,
            show_path: true,
            show_spans: true,
            style,
            color_depth: ColorDepth::TrueColor,
        }
    }

    /// Returns a copy of the current style configuration.
    #[must_use]
    pub fn style_config(&self) -> Style {
        **self.style.load()
    }

    /// Returns a clone of the internal style handle for sharing with TracingGuard.
    /// When TracingGuard modifies the handle (via `with_style`), all Formatters
    /// sharing this handle will see the update on the next format_event call.
    #[must_use]
    pub fn style_handle(&self) -> Arc<arc_swap::ArcSwap<Style>> {
        self.style.clone()
    }

    #[must_use]
    pub fn with_style_config(self, new_style: Style) -> Self {
        self.style.store(Arc::new(new_style));
        self
    }

    #[must_use]
    pub fn with_icons(self, icons: Icons) -> Self {
        let mut style = **self.style.load();
        style.icons = icons;
        self.style.store(Arc::new(style));
        self
    }

    #[must_use]
    pub fn with_labels(self, labels: LevelLabels) -> Self {
        let mut style = **self.style.load();
        style.labels = labels;
        self.style.store(Arc::new(style));
        self
    }

    #[must_use]
    pub fn with_theme(self, theme: Theme) -> Self {
        let mut style = **self.style.load();
        style.theme = theme;
        self.style.store(Arc::new(style));
        self
    }

    #[must_use]
    pub const fn with_color_depth(mut self, depth: ColorDepth) -> Self {
        self.color_depth = depth;
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

    pub(crate) fn format_path(&self, file: &str, line: u32) -> ArrayString<PATH_BUF_SIZE> {
        let path_start = file.find("src/").or_else(|| file.find("src\\"));
        let path = path_start.map_or(file, |i| file.get(i.saturating_add(4)..).unwrap_or(file));

        let mut norm_path = ArrayString::<PATH_BUF_SIZE>::new();
        for c in path.chars() {
            if norm_path.len() < norm_path.capacity() {
                let replaced = if c == '\\' { '/' } else { c };
                norm_path.push(replaced);
            }
        }
        let path_str = norm_path.as_str();

        let max_width = self.path_width;
        let mut full = ArrayString::<PATH_BUF_SIZE>::new();
        let _ = write!(full, "{path_str}:{line}");

        if full.len() <= max_width {
            let mut result = ArrayString::<PATH_BUF_SIZE>::new();
            write!(result, "{full:>max_width$}").ok();
            return result;
        }

        if let Some(last_slash) = path_str.rfind('/') {
            let tail = path_str
                .get(last_slash.saturating_add(1)..)
                .unwrap_or(path_str);
            let mut file_part = ArrayString::<PATH_BUF_SIZE>::new();
            let _ = write!(file_part, "{tail}:{line}");

            if file_part.len().saturating_add(2) <= max_width {
                let dir_part = path_str.get(..last_slash).unwrap_or("");
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

    fn write_time(&self, writer: &mut Writer<'_>, theme: &Theme) -> fmt::Result {
        let now = Utc::now();
        write!(
            writer,
            "{}",
            rgb_to_owo(theme.text, self.color_depth).style(now.format(&self.time_format))
        )
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
            theme_fg_dimmed(theme.text, self.color_depth).style(self.format_path(
                event.metadata().file().unwrap_or("?"),
                event.metadata().line().unwrap_or(0),
            ))
        )?;
        write!(
            writer,
            " {} ",
            rgb_to_owo(theme.accent, self.color_depth).style(icons.arrow)
        )?;
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

        let mut sep = if let Some(msg) = visitor.message {
            write!(
                writer,
                "{}",
                rgb_to_owo(theme.text, self.color_depth).style(msg)
            )?;
            " "
        } else {
            ""
        };

        for (k, v) in visitor.fields {
            write!(
                writer,
                "{}{}{}{}",
                sep,
                rgb_to_owo(theme.secondary, self.color_depth).style(k),
                rgb_to_owo(theme.accent, self.color_depth).style("="),
                rgb_to_owo(theme.text, self.color_depth).style(v)
            )?;
            sep = " ";
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
        let accent = rgb_to_owo(theme.accent, self.color_depth);
        let accent_dimmed = theme_fg_dimmed(theme.accent, self.color_depth);
        let text = rgb_to_owo(theme.text, self.color_depth);
        let text_dimmed = theme_fg_dimmed(theme.text, self.color_depth);

        write!(writer, " {}", accent.style("["))?;

        for (i, span) in spans.iter().enumerate() {
            if i > 0 {
                write!(writer, "{} ", accent_dimmed.style(icons.span_join))?;
            }

            let span_style = if i == total - 1 { text } else { text_dimmed };

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

impl<S, N> FormatEvent<S, N> for Formatter
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
        let config = self.style.load();

        let level = event.metadata().level();

        let (color, level_label) = match *level {
            Level::ERROR => (config.theme.error, config.labels.error),
            Level::WARN => (config.theme.warn, config.labels.warn),
            Level::INFO => (config.theme.info, config.labels.info),
            Level::DEBUG => (config.theme.debug, config.labels.debug),
            Level::TRACE => (config.theme.trace, config.labels.trace),
        };

        let ls = {
            let r = color.0;
            let g = color.1;
            let b = color.2;
            let depth = self.color_depth;
            LevelStyles {
                bracket_fg: rgb_to_owo((r, g, b), depth),
                bracket_bg: rgb_to_owo_on(r, g, b, depth),
                label: rgb_to_owo_on(r, g, b, depth),
            }
        };

        let fg_style = match config.icons.name {
            "nerd" => ls.bracket_fg,
            _ => ls.bracket_bg,
        };

        write!(
            writer,
            "{}",
            rgb_to_owo(config.theme.accent, self.color_depth).style(config.icons.time_bracket_open)
        )?;
        self.write_time(&mut writer, &config.theme)?;
        write!(
            writer,
            " {} ",
            theme_fg_dimmed(config.theme.accent, self.color_depth).style(config.icons.separator)
        )?;

        write!(writer, "{}", fg_style.style(config.icons.bracket_open))?;
        write!(writer, "{}", ls.label.style(level_label))?;
        write!(writer, "{} ", fg_style.style(config.icons.bracket_close))?;

        write!(
            writer,
            "{} ",
            rgb_to_owo(config.theme.accent, self.color_depth)
                .style(config.icons.time_bracket_close)
        )?;

        if self.show_path {
            self.format_path_section(&mut writer, event, &config.theme, &config.icons)?;
        }

        self.format_fields(&mut writer, event, &config.theme)?;

        if self.show_spans {
            self.format_spans(&mut writer, ctx, &config.theme, &config.icons)?;
        }

        writeln!(writer)
    }
}

#[cfg(test)]
mod test;
