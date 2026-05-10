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
    pub path: Option<PathBuf>,
    pub rotation: LogRotation,
}

impl Default for FileLoggingConfigBuilder {
    fn default() -> Self {
        Self {
            path: None,
            rotation: LogRotation::default(),
        }
    }
}

impl FileLoggingConfigBuilder {
    pub fn path(mut self, path: impl Into<PathBuf>) -> Self {
        self.path = Some(path.into());
        self
    }

    pub fn rotation(mut self, rotation: LogRotation) -> Self {
        self.rotation = rotation;
        self
    }

    pub fn build(self) -> FileLoggingConfig {
        FileLoggingConfig {
            path: self.path.unwrap_or_else(|| PathBuf::from("app.log")),
            rotation: self.rotation,
        }
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
}
impl ConsoleConfig {
    pub fn builder() -> ConsoleConfigBuilder {
        ConsoleConfigBuilder::default()
    }
}

#[derive(Clone, Debug)]
pub struct ConsoleConfigBuilder {
    format: LogFormat,
    ansi: bool,
    writer: ConsoleWriter,
    show_path: bool,
    show_spans: bool,
    time_format: Option<String>,
}

impl Default for ConsoleConfigBuilder {
    fn default() -> Self {
        Self {
            format: LogFormat::default(),
            ansi: true,
            writer: ConsoleWriter::default(),
            show_path: true,
            show_spans: true,
            time_format: None,
        }
    }
}

impl ConsoleConfigBuilder {
    pub fn format(mut self, format: LogFormat) -> Self {
        self.format = format;
        self
    }

    pub fn ansi(mut self, ansi: bool) -> Self {
        self.ansi = ansi;
        self
    }

    pub fn writer(mut self, writer: ConsoleWriter) -> Self {
        self.writer = writer;
        self
    }

    pub fn show_path(mut self, show: bool) -> Self {
        self.show_path = show;
        self
    }

    pub fn show_spans(mut self, show: bool) -> Self {
        self.show_spans = show;
        self
    }

    pub fn time_format(mut self, fmt: impl Into<String>) -> Self {
        self.time_format = Some(fmt.into());
        self
    }

    pub fn build(self) -> ConsoleConfig {
        ConsoleConfig {
            format: self.format,
            ansi: self.ansi,
            writer: self.writer,
            show_path: self.show_path,
            show_spans: self.show_spans,
            time_format: self.time_format,
        }
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

#[derive(Clone, Debug)]
pub struct LoggingConfigBuilder {
    level: LogLevel,
    console: Option<ConsoleConfig>,
    file: Option<FileLoggingConfig>,
}

impl Default for LoggingConfigBuilder {
    fn default() -> Self {
        Self {
            level: LogLevel::Info,
            console: Some(ConsoleConfig::default()),
            file: None,
        }
    }
}

impl LoggingConfigBuilder {
    pub fn level(mut self, level: LogLevel) -> Self {
        self.level = level;
        self
    }

    pub fn console(mut self, console: ConsoleConfig) -> Self {
        self.console = Some(console);
        self
    }

    pub fn no_console(mut self) -> Self {
        self.console = None;
        self
    }

    pub fn file(mut self, file: FileLoggingConfig) -> Self {
        self.file = Some(file);
        self
    }

    pub fn build(self) -> LoggingConfig {
        LoggingConfig {
            level: self.level,
            console: self.console,
            file: self.file,
        }
    }
}

#[cfg(feature = "serde")]
fn default_console() -> Option<ConsoleConfig> {
    Some(ConsoleConfig::default())
}

#[cfg(test)]
mod test;
