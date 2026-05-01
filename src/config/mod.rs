use std::path::PathBuf;

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "lowercase"))]
#[derive(Clone, Debug)]
#[non_exhaustive]
pub enum LogFormat {
    Pretty,
    Compact,
    Json,
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "lowercase"))]
#[derive(Clone, Debug, Default)]
#[non_exhaustive]
pub enum LogRotation {
    #[default]
    None,
    Rename,
    Compress,
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
    Custom(String),
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
            Self::Custom(directive) => directive,
        }
    }
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[derive(Clone, Debug)]
pub struct LogFilter {
    pub level: LogLevel,
    pub targets: Vec<(String, LogLevel)>,
}

impl LogFilter {
    pub fn new(level: LogLevel) -> Self {
        Self {
            level,
            targets: Vec::new(),
        }
    }

    pub fn with_target_level(mut self, target: impl Into<String>, level: LogLevel) -> Self {
        self.set_target_level(target, level);
        self
    }

    pub fn set_target_level(&mut self, target: impl Into<String>, level: LogLevel) {
        let target = target.into();
        if let Some((_, existing)) = self.targets.iter_mut().find(|(name, _)| name == &target) {
            *existing = level;
        } else {
            self.targets.push((target, level));
        }
    }

    pub fn remove_target_level(&mut self, target: &str) -> bool {
        let len = self.targets.len();
        self.targets.retain(|(name, _)| name != target);
        self.targets.len() != len
    }

    pub fn as_filter_directive(&self) -> String {
        let mut directive = self.level.as_filter_directive().to_string();
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
            other => Self::Custom(other.to_string()),
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

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "lowercase"))]
#[derive(Clone, Debug, Default)]
#[non_exhaustive]
pub enum ConsoleWriter {
    #[default]
    Stdout,
    Stderr,
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[derive(Clone, Debug)]
pub struct ConsoleConfig {
    pub format: LogFormat,
    #[cfg_attr(feature = "serde", serde(default = "default_true"))]
    pub ansi: bool,
    #[cfg_attr(feature = "serde", serde(default))]
    pub writer: ConsoleWriter,
    #[cfg_attr(feature = "serde", serde(default = "default_true"))]
    pub show_path: bool,
    #[cfg_attr(feature = "serde", serde(default = "default_true"))]
    pub show_spans: bool,
    #[cfg_attr(feature = "serde", serde(default))]
    pub time_format: Option<String>,
}

#[cfg(feature = "serde")]
fn default_true() -> bool {
    true
}

impl Default for ConsoleConfig {
    fn default() -> Self {
        Self {
            format: LogFormat::Compact,
            ansi: true,
            writer: ConsoleWriter::default(),
            show_path: true,
            show_spans: true,
            time_format: None,
        }
    }
}

#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[derive(Clone, Debug)]
pub struct LoggingConfig {
    pub level: LogLevel,
    #[cfg_attr(feature = "serde", serde(default = "default_console"))]
    pub console: Option<ConsoleConfig>,
    #[cfg_attr(feature = "serde", serde(default))]
    pub file: Option<FileLoggingConfig>,
}

#[cfg(feature = "serde")]
fn default_console() -> Option<ConsoleConfig> {
    Some(ConsoleConfig::default())
}

impl Default for LoggingConfig {
    fn default() -> Self {
        Self {
            level: LogLevel::Info,
            console: Some(ConsoleConfig::default()),
            file: None,
        }
    }
}

#[cfg(test)]
mod test;
