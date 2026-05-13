#[cfg(feature = "nerd")]
use nerd_font_symbols::{cod, fa, ple};
use owo_colors::Style;
use smart_default::SmartDefault;
use std::collections::HashMap;
use std::path::PathBuf;

#[derive(Clone, Copy, Debug, PartialEq)]
#[non_exhaustive]
pub struct LevelLabels {
    pub error: &'static str,
    pub warn: &'static str,
    pub info: &'static str,
    pub debug: &'static str,
    pub trace: &'static str,
}

impl LevelLabels {
    pub const fn custom(
        error: &'static str,
        warn: &'static str,
        info: &'static str,
        debug: &'static str,
        trace: &'static str,
    ) -> Self {
        Self {
            error,
            warn,
            info,
            debug,
            trace,
        }
    }
    pub const fn short() -> Self {
        Self {
            error: "E",
            warn: "W",
            info: "I",
            debug: "D",
            trace: "T",
        }
    }
    pub const fn long() -> Self {
        Self {
            error: "ERROR",
            warn: " WARN",
            info: " INFO",
            debug: "DEBUG",
            trace: "TRACE",
        }
    }
}

impl Default for LevelLabels {
    fn default() -> Self {
        Self::short()
    }
}

#[derive(Clone, Copy, Debug, PartialEq)]
#[non_exhaustive]
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
    #[allow(clippy::too_many_arguments)]
    pub const fn custom(
        bracket_open: &'static str,
        bracket_close: &'static str,
        time_bracket_open: &'static str,
        time_bracket_close: &'static str,
        separator: &'static str,
        arrow: &'static str,
        span_delimiter: &'static str,
        span_join: &'static str,
    ) -> Self {
        Self {
            bracket_open,
            bracket_close,
            time_bracket_open,
            time_bracket_close,
            separator,
            arrow,
            span_delimiter,
            span_join,
        }
    }

    pub const fn unicode() -> Self {
        Self {
            bracket_open: "[",
            bracket_close: "]",
            time_bracket_open: "\u{300c}",
            time_bracket_close: "\u{300d}",
            separator: "\u{2507}",
            arrow: ">",
            span_delimiter: "->",
            span_join: "\u{00bb}",
        }
    }

    #[cfg(feature = "nerd")]
    pub const fn nerd() -> Self {
        Self {
            bracket_open: ple::PLE_LEFT_HALF_CIRCLE_THICK,
            bracket_close: ple::PLE_RIGHT_HALF_CIRCLE_THICK,
            time_bracket_open: ple::PLE_LEFT_HALF_CIRCLE_THIN,
            time_bracket_close: ple::PLE_RIGHT_HALF_CIRCLE_THIN,
            separator: "\u{2507}",
            arrow: fa::FA_CARET_RIGHT,
            span_delimiter: cod::COD_EXPORT,
            span_join: fa::FA_ANGLES_RIGHT,
        }
    }

    pub fn is_nerd(&self) -> bool {
        self.bracket_open != "["
    }
}

impl Default for Icons {
    fn default() -> Self {
        #[cfg(feature = "nerd")]
        {
            Self::nerd()
        }
        #[cfg(not(feature = "nerd"))]
        {
            Self::unicode()
        }
    }
}

type Rgb = (u8, u8, u8);

#[derive(Clone, Copy, Debug)]
#[non_exhaustive]
pub struct Theme {
    pub accent: Style,
    pub secondary: Style,
    pub text: Style,
    pub error: Rgb,
    pub warn: Rgb,
    pub info: Rgb,
    pub debug: Rgb,
    pub trace: Rgb,
}

impl Theme {
    #[allow(clippy::too_many_arguments)]
    pub fn new(
        accent: Rgb,
        secondary: Rgb,
        text: Rgb,
        error: Rgb,
        warn: Rgb,
        info: Rgb,
        debug: Rgb,
        trace: Rgb,
    ) -> Self {
        Self {
            accent: Style::new().truecolor(accent.0, accent.1, accent.2),
            secondary: Style::new().truecolor(secondary.0, secondary.1, secondary.2),
            text: Style::new().truecolor(text.0, text.1, text.2),
            error,
            warn,
            info,
            debug,
            trace,
        }
    }

    pub fn trans_flag() -> Self {
        Self::new(
            (91, 206, 250),
            (245, 169, 184),
            (255, 255, 255),
            (255, 85, 85),
            (255, 200, 60),
            (91, 206, 250),
            (245, 169, 184),
            (240, 240, 240),
        )
    }
    pub fn monokai() -> Self {
        Self::new(
            (102, 217, 239),
            (249, 38, 114),
            (248, 248, 242),
            (255, 85, 85),
            (255, 200, 60),
            (102, 217, 239),
            (249, 38, 114),
            (180, 180, 180),
        )
    }
    pub fn dracula() -> Self {
        Self::new(
            (139, 233, 253),
            (255, 121, 198),
            (248, 248, 242),
            (255, 85, 85),
            (255, 200, 60),
            (139, 233, 253),
            (255, 121, 198),
            (180, 180, 180),
        )
    }
    pub fn nord() -> Self {
        Self::new(
            (136, 192, 208),
            (163, 190, 140),
            (216, 222, 233),
            (191, 97, 106),
            (235, 203, 139),
            (136, 192, 208),
            (163, 190, 140),
            (180, 180, 180),
        )
    }
    pub fn catppuccin_mocha() -> Self {
        Self::new(
            (137, 180, 250),
            (203, 166, 247),
            (205, 214, 244),
            (243, 139, 168),
            (249, 226, 175),
            (137, 180, 250),
            (203, 166, 247),
            (180, 180, 180),
        )
    }
    pub fn gruvbox() -> Self {
        Self::new(
            (131, 165, 152),
            (254, 128, 25),
            (235, 219, 178),
            (251, 73, 52),
            (250, 189, 47),
            (131, 165, 152),
            (254, 128, 25),
            (180, 180, 180),
        )
    }
    pub fn one_dark() -> Self {
        Self::new(
            (97, 175, 239),
            (198, 120, 221),
            (171, 178, 191),
            (224, 108, 117),
            (229, 192, 123),
            (97, 175, 239),
            (198, 120, 221),
            (180, 180, 180),
        )
    }
    pub fn tokyo_night() -> Self {
        Self::new(
            (122, 162, 247),
            (187, 154, 247),
            (192, 202, 245),
            (247, 118, 142),
            (224, 175, 104),
            (122, 162, 247),
            (187, 154, 247),
            (180, 180, 180),
        )
    }
}

impl Default for Theme {
    fn default() -> Self {
        Self::trans_flag()
    }
}

#[derive(Clone, Copy, Debug, Default)]
#[non_exhaustive]
pub struct StyleConfig {
    pub theme: Theme,
    pub icons: Icons,
    pub labels: LevelLabels,
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "lowercase"))]
#[derive(Clone, Debug, Default)]
#[non_exhaustive]
pub enum Format {
    Pretty,
    #[default]
    Compact,
    Json,
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "lowercase"))]
#[derive(Clone, Copy, Debug, Default)]
#[non_exhaustive]
pub enum Rotation {
    #[default]
    None,
    Rename,
    #[cfg(feature = "compress")]
    Compress,
}

#[derive(Clone, Debug)]
#[non_exhaustive]
pub enum Level {
    Error,
    Warn,
    Info,
    Debug,
    Trace,
    Off,
    Custom(String),
}

impl Level {
    pub fn as_directive(&self) -> &str {
        match self {
            Self::Error => "error",
            Self::Warn => "warn",
            Self::Info => "info",
            Self::Debug => "debug",
            Self::Trace => "trace",
            Self::Off => "off",
            Self::Custom(s) => s.as_str(),
        }
    }
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[derive(Clone, Debug)]
pub struct Filter {
    pub level: Level,
    pub targets: HashMap<String, Level>,
}

impl Filter {
    pub fn new(level: Level) -> Self {
        Self {
            level,
            targets: HashMap::new(),
        }
    }

    pub fn with_target(mut self, target: impl Into<String>, level: Level) -> Self {
        self.set_target(target, level);
        self
    }

    pub fn set_target(&mut self, target: impl Into<String>, level: Level) {
        self.targets.insert(target.into(), level);
    }

    pub fn remove_target(&mut self, target: &str) -> bool {
        self.targets.remove(target).is_some()
    }

    pub fn as_directive(&self) -> String {
        let mut directive = String::from(self.level.as_directive());
        for (target, level) in &self.targets {
            directive.push(',');
            directive.push_str(target);
            directive.push('=');
            directive.push_str(level.as_directive());
        }
        directive
    }
}

impl From<Level> for Filter {
    fn from(level: Level) -> Self {
        Self::new(level)
    }
}

#[cfg(feature = "serde")]
impl serde::Serialize for Level {
    fn serialize<S: serde::Serializer>(&self, serializer: S) -> Result<S::Ok, S::Error> {
        serializer.serialize_str(self.as_directive())
    }
}

#[cfg(feature = "serde")]
impl<'de> serde::Deserialize<'de> for Level {
    fn deserialize<D: serde::Deserializer<'de>>(deserializer: D) -> Result<Self, D::Error> {
        let s = String::deserialize(deserializer)?;
        Ok(match s.as_str() {
            "error" => Self::Error,
            "warn" => Self::Warn,
            "info" => Self::Info,
            "debug" => Self::Debug,
            "trace" => Self::Trace,
            "off" => Self::Off,
            other => Self::Custom(other.to_owned()),
        })
    }
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[derive(Clone, Debug)]
pub struct FileConfig {
    pub path: PathBuf,
    #[cfg_attr(feature = "serde", serde(default))]
    pub rotation: Rotation,
}

impl Default for FileConfig {
    fn default() -> Self {
        Self {
            path: PathBuf::from("app.log"),
            rotation: Rotation::default(),
        }
    }
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "lowercase"))]
#[derive(Clone, Debug, Default)]
#[non_exhaustive]
pub enum Writer {
    #[default]
    Stdout,
    Stderr,
    #[cfg(any(feature = "custom-async", feature = "native-async"))]
    AsyncStdout(AsyncMode),
    #[cfg(any(feature = "custom-async", feature = "native-async"))]
    AsyncStderr(AsyncMode),
}

#[cfg(any(feature = "custom-async", feature = "native-async"))]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "lowercase"))]
#[derive(Clone, Copy, Debug)]
pub enum AsyncMode {
    #[cfg(feature = "custom-async")]
    Custom,
    #[cfg(feature = "native-async")]
    Native,
}

#[cfg(any(feature = "custom-async", feature = "native-async"))]
#[allow(clippy::derivable_impls)]
impl Default for AsyncMode {
    fn default() -> Self {
        #[cfg(feature = "custom-async")]
        {
            Self::Custom
        }
        #[cfg(all(feature = "native-async", not(feature = "custom-async")))]
        {
            Self::Native
        }
    }
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[derive(Clone, Debug, SmartDefault)]
pub struct ConsoleConfig {
    #[default(Format::default())]
    pub format: Format,
    #[default = true]
    #[cfg_attr(feature = "serde", serde(default = "default_true"))]
    pub ansi: bool,
    #[cfg_attr(feature = "serde", serde(default))]
    pub writer: Writer,
    #[default = true]
    #[cfg_attr(feature = "serde", serde(default = "default_true"))]
    pub show_path: bool,
    #[default = true]
    #[cfg_attr(feature = "serde", serde(default = "default_true"))]
    pub show_spans: bool,
    #[cfg_attr(feature = "serde", serde(default))]
    pub time_format: Option<String>,
    #[default(StyleConfig::default())]
    #[cfg_attr(feature = "serde", serde(skip))]
    pub style: StyleConfig,
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[derive(Clone, Debug, SmartDefault)]
pub struct LoggingConfig {
    #[default(Level::Info)]
    pub level: Level,
    #[default(Some(ConsoleConfig::default()))]
    #[cfg_attr(feature = "serde", serde(default = "default_console"))]
    pub console: Option<ConsoleConfig>,
    #[cfg_attr(feature = "serde", serde(default))]
    pub file: Option<FileConfig>,
}

#[cfg(feature = "serde")]
fn default_true() -> bool {
    true
}

#[cfg(feature = "serde")]
fn default_console() -> Option<ConsoleConfig> {
    Some(ConsoleConfig::default())
}

#[cfg(test)]
mod test;
