use crate::fmt::StyleConfig;
use smart_default::SmartDefault;
use std::collections::HashMap;
use std::path::PathBuf;
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "lowercase"))]
#[derive(Clone, Debug, Default)]
#[non_exhaustive]
pub enum LogFormat {
    Pretty,
    #[default]
    Compact,
    Json,
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "lowercase"))]
#[derive(Clone, Copy, Debug, Default)]
#[non_exhaustive]
pub enum LogRotation {
    #[default]
    None,
    Rename,
    #[cfg(feature = "compress")]
    Compress,
}
#[derive(Clone, Debug, PartialEq, Eq, Hash, derive_more::Display, derive_more::From)]
pub struct FilterDirective(String);

impl FilterDirective {
    pub fn new(value: impl Into<String>) -> Self {
        Self(value.into())
    }

    pub fn as_str(&self) -> &str {
        &self.0
    }
}

impl AsRef<str> for FilterDirective {
    fn as_ref(&self) -> &str {
        &self.0
    }
}

#[derive(Clone, Debug)]
#[non_exhaustive]
pub enum LogLevel {
    Error,
    Warn,
    Info,
    Debug,
    Trace,
    Off,
    Custom(FilterDirective),
}

impl LogLevel {
    pub fn as_filter_directive(&self) -> &str {
        match self {
            Self::Error => "error",
            Self::Warn => "warn",
            Self::Info => "info",
            Self::Debug => "debug",
            Self::Trace => "trace",
            Self::Off => "off",
            Self::Custom(directive) => directive.as_str(),
        }
    }
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[derive(Clone, Debug)]
pub struct LogFilter {
    pub level: LogLevel,
    pub targets: HashMap<String, LogLevel>,
}

impl LogFilter {
    pub fn new(level: LogLevel) -> Self {
        Self {
            level,
            targets: HashMap::new(),
        }
    }

    pub fn with_target_level(mut self, target: impl Into<String>, level: LogLevel) -> Self {
        self.set_target_level(target, level);
        self
    }

    pub fn set_target_level(&mut self, target: impl Into<String>, level: LogLevel) {
        self.targets.insert(target.into(), level);
    }

    pub fn remove_target_level(&mut self, target: &str) -> bool {
        self.targets.remove(target).is_some()
    }

    pub fn as_filter_directive(&self) -> String {
        let mut directive = String::from(self.level.as_filter_directive());
        for (target, level) in &self.targets {
            directive.push(',');
            directive.push_str(target);
            directive.push('=');
            directive.push_str(level.as_filter_directive());
        }
        directive
    }
}

impl From<LogLevel> for LogFilter {
    fn from(level: LogLevel) -> Self {
        Self::new(level)
    }
}

#[cfg(feature = "serde")]
impl serde::Serialize for LogLevel {
    fn serialize<S: serde::Serializer>(&self, serializer: S) -> Result<S::Ok, S::Error> {
        serializer.serialize_str(self.as_filter_directive())
    }
}

#[cfg(feature = "serde")]
impl<'de> serde::Deserialize<'de> for LogLevel {
    fn deserialize<D: serde::Deserializer<'de>>(deserializer: D) -> Result<Self, D::Error> {
        let s = String::deserialize(deserializer)?;
        Ok(match s.as_str() {
            "error" => Self::Error,
            "warn" => Self::Warn,
            "info" => Self::Info,
            "debug" => Self::Debug,
            "trace" => Self::Trace,
            "off" => Self::Off,
            other => Self::Custom(FilterDirective::new(other)),
        })
    }
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[derive(Clone, Debug)]
pub struct FileLoggingConfig {
    pub path: PathBuf,
    #[cfg_attr(feature = "serde", serde(default))]
    pub rotation: LogRotation,
}
impl FileLoggingConfig {
    pub fn builder() -> FileLoggingConfigBuilder {
        FileLoggingConfigBuilder::default()
    }
}

#[derive(Clone, Debug)]
pub struct FileLoggingConfigBuilder {
    inner: FileLoggingConfig,
}

impl Default for FileLoggingConfigBuilder {
    fn default() -> Self {
        Self {
            inner: FileLoggingConfig {
                path: PathBuf::from("app.log"),
                rotation: LogRotation::default(),
            },
        }
    }
}

impl FileLoggingConfigBuilder {
    pub fn path(mut self, path: impl Into<PathBuf>) -> Self {
        self.inner.path = path.into();
        self
    }

    pub fn rotation(mut self, rotation: LogRotation) -> Self {
        self.inner.rotation = rotation;
        self
    }

    pub fn build(self) -> FileLoggingConfig {
        self.inner
    }
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "lowercase"))]
#[derive(Clone, Debug, Default)]
#[non_exhaustive]
pub enum ConsoleWriter {
    #[default]
    Stdout,
    Stderr,
    #[cfg(any(feature = "custom-async", feature = "native-async"))]
    AsyncStdout(AsyncWriterMode),
    #[cfg(any(feature = "custom-async", feature = "native-async"))]
    AsyncStderr(AsyncWriterMode),
}

#[cfg(any(feature = "custom-async", feature = "native-async"))]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "lowercase"))]
#[derive(Clone, Copy, Debug, Default)]
pub enum AsyncWriterMode {
    #[cfg(feature = "custom-async")]
    #[default]
    Custom,
    #[cfg(feature = "native-async")]
    Native,
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[derive(Clone, Debug, SmartDefault)]
pub struct ConsoleConfig {
    #[default(LogFormat::default())]
    pub format: LogFormat,
    #[default = true]
    #[cfg_attr(feature = "serde", serde(default = "default_true"))]
    pub ansi: bool,
    #[cfg_attr(feature = "serde", serde(default))]
    pub writer: ConsoleWriter,
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
impl ConsoleConfig {
    pub fn builder() -> ConsoleConfigBuilder {
        ConsoleConfigBuilder::default()
    }
}

#[derive(Clone, Debug, Default)]
pub struct ConsoleConfigBuilder {
    inner: ConsoleConfig,
}

impl ConsoleConfigBuilder {
    pub fn format(mut self, format: LogFormat) -> Self {
        self.inner.format = format;
        self
    }

    pub fn ansi(mut self, ansi: bool) -> Self {
        self.inner.ansi = ansi;
        self
    }

    pub fn writer(mut self, writer: ConsoleWriter) -> Self {
        self.inner.writer = writer;
        self
    }

    pub fn show_path(mut self, show: bool) -> Self {
        self.inner.show_path = show;
        self
    }

    pub fn show_spans(mut self, show: bool) -> Self {
        self.inner.show_spans = show;
        self
    }

    pub fn time_format(mut self, fmt: impl Into<String>) -> Self {
        self.inner.time_format = Some(fmt.into());
        self
    }

    pub fn style(mut self, style: StyleConfig) -> Self {
        self.inner.style = style;
        self
    }

    pub fn build(self) -> ConsoleConfig {
        self.inner
    }
}

#[cfg(feature = "serde")]
fn default_true() -> bool {
    true
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[derive(Clone, Debug, SmartDefault)]
pub struct LoggingConfig {
    #[default(LogLevel::Info)]
    pub level: LogLevel,
    #[default(Some(ConsoleConfig::default()))]
    #[cfg_attr(feature = "serde", serde(default = "default_console"))]
    pub console: Option<ConsoleConfig>,
    #[cfg_attr(feature = "serde", serde(default))]
    pub file: Option<FileLoggingConfig>,
}
impl LoggingConfig {
    pub fn builder() -> LoggingConfigBuilder {
        LoggingConfigBuilder::default()
    }
}

#[derive(Clone, Debug, Default)]
pub struct LoggingConfigBuilder {
    inner: LoggingConfig,
}

impl LoggingConfigBuilder {
    pub fn level(mut self, level: LogLevel) -> Self {
        self.inner.level = level;
        self
    }

    pub fn console(mut self, console: ConsoleConfig) -> Self {
        self.inner.console = Some(console);
        self
    }

    pub fn no_console(mut self) -> Self {
        self.inner.console = None;
        self
    }

    pub fn file(mut self, file: FileLoggingConfig) -> Self {
        self.inner.file = Some(file);
        self
    }

    pub fn build(self) -> LoggingConfig {
        self.inner
    }
}

#[cfg(feature = "serde")]
fn default_console() -> Option<ConsoleConfig> {
    Some(ConsoleConfig::default())
}

#[cfg(test)]
mod test;
