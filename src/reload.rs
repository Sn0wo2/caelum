use std::sync::{Arc, RwLock};

use tracing_subscriber::EnvFilter;
use tracing_subscriber::layer::Layered;

use crate::config::{FilterDirective, LogFilter, LogLevel};
use crate::error::{Result, SageTraceError};
use crate::fmt::{Icons, LevelLabels, StyleConfig, Theme};

pub(crate) type FmtLayer =
    Box<dyn tracing_subscriber::Layer<tracing_subscriber::Registry> + Send + Sync>;
pub(crate) type InnerSubscriber = Layered<FmtLayer, tracing_subscriber::Registry>;
type RawReloadHandle = tracing_subscriber::reload::Handle<EnvFilter, InnerSubscriber>;

#[must_use = "dropping ReloadHandle loses the ability to change log filters at runtime"]
#[derive(Clone)]
pub struct ReloadHandle {
    raw: RawReloadHandle,
    filter: Arc<RwLock<LogFilter>>,
    style: Option<Arc<RwLock<StyleConfig>>>,
}

impl std::fmt::Debug for ReloadHandle {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("ReloadHandle").finish_non_exhaustive()
    }
}

impl ReloadHandle {
    pub fn reload(&self, directive: &str) -> Result<()> {
        self.apply_directive(directive)?;
        self.store_filter(LogFilter::new(LogLevel::Custom(FilterDirective::new(
            directive,
        ))))?;
        Ok(())
    }

    pub fn set_filter(&self, filter: LogFilter) -> Result<()> {
        let directive = filter.as_filter_directive();
        self.apply_directive(&directive)?;
        self.store_filter(filter)?;
        Ok(())
    }

    pub fn set_level(&self, level: LogLevel) -> Result<()> {
        self.update_filter(|filter| filter.level = level)
    }

    pub fn set_target_level(&self, target: impl Into<String>, level: LogLevel) -> Result<()> {
        let target = target.into();
        self.update_filter(|filter| filter.set_target_level(target, level))
    }

    pub fn remove_target_level(&self, target: &str) -> Result<()> {
        self.update_filter(|filter| {
            filter.remove_target_level(target);
        })
    }

    fn with_style(&self, f: impl FnOnce(&mut StyleConfig)) -> Result<()> {
        let style = self.style.as_ref().ok_or(SageTraceError::StyleNotConfigured)?;
        let mut guard = style.write().map_err(|_| SageTraceError::LockPoisoned)?;
        f(&mut guard);
        Ok(())
    }

    pub fn set_icons(&self, icons: Icons) -> Result<()> {
        self.with_style(|s| s.icons = icons)
    }

    pub fn set_theme(&self, theme: Theme) -> Result<()> {
        self.with_style(|s| s.theme = theme)
    }

    pub fn set_labels(&self, labels: LevelLabels) -> Result<()> {
        self.with_style(|s| s.labels = labels)
    }

    fn update_filter(&self, update: impl FnOnce(&mut LogFilter)) -> Result<()> {
        let mut next = self.current_filter()?;
        update(&mut next);
        self.set_filter(next)
    }

    fn current_filter(&self) -> Result<LogFilter> {
        Ok(self
            .filter
            .read()
            .map_err(|_| SageTraceError::LockPoisoned)?
            .clone())
    }

    fn store_filter(&self, filter: LogFilter) -> Result<()> {
        *self
            .filter
            .write()
            .map_err(|_| SageTraceError::LockPoisoned)? = filter;
        Ok(())
    }

    fn apply_directive(&self, directive: &str) -> Result<()> {
        let filter = EnvFilter::try_new(directive)?;
        self.raw.modify(|f| *f = filter)?;
        Ok(())
    }
}

pub fn build_reload_filter(
    level: &LogLevel,
    style: Option<Arc<RwLock<StyleConfig>>>,
) -> (
    tracing_subscriber::reload::Layer<EnvFilter, InnerSubscriber>,
    ReloadHandle,
) {
    let filter = EnvFilter::new(level.as_filter_directive());
    let (layer, raw_handle) = tracing_subscriber::reload::Layer::new(filter);
    (
        layer,
        ReloadHandle {
            raw: raw_handle,
            filter: Arc::new(RwLock::new(LogFilter::new(level.clone()))),
            style,
        },
    )
}
