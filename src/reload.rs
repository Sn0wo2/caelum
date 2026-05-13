use tracing_subscriber::EnvFilter;
use tracing_subscriber::layer::Layered;

use crate::config::{FilterDirective, LogFilter, LogLevel};
use crate::error::Result;
use crate::fmt::StyleConfig;

pub(crate) type FmtLayer =
    Box<dyn tracing_subscriber::Layer<tracing_subscriber::Registry> + Send + Sync>;
pub(crate) type InnerSubscriber = Layered<FmtLayer, tracing_subscriber::Registry>;
type RawReloadHandle = tracing_subscriber::reload::Handle<EnvFilter, InnerSubscriber>;

#[must_use = "dropping ReloadHandle loses the ability to change log filters at runtime"]
pub struct ReloadHandle {
    raw: RawReloadHandle,
    filter: LogFilter,
    style: StyleConfig,
}

impl std::fmt::Debug for ReloadHandle {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("ReloadHandle").finish_non_exhaustive()
    }
}

impl ReloadHandle {
    pub fn with_style(&mut self, f: impl FnOnce(&mut StyleConfig)) {
        f(&mut self.style);
    }

    pub fn reload(&mut self, directive: &str) -> Result<()> {
        self.apply_directive(directive)?;
        self.filter = LogFilter::new(LogLevel::Custom(FilterDirective::new(directive)));
        Ok(())
    }

    pub fn set_filter(&mut self, filter: LogFilter) -> Result<()> {
        let directive = filter.as_filter_directive();
        self.apply_directive(&directive)?;
        self.filter = filter;
        Ok(())
    }

    pub fn set_level(&mut self, level: LogLevel) -> Result<()> {
        self.filter.level = level;
        self.apply_current_filter()
    }

    pub fn set_target_level(&mut self, target: impl Into<String>, level: LogLevel) -> Result<()> {
        let target = target.into();
        self.filter.set_target_level(target, level);
        self.apply_current_filter()
    }

    pub fn remove_target_level(&mut self, target: &str) -> Result<()> {
        self.filter.remove_target_level(target);
        self.apply_current_filter()
    }

    fn apply_current_filter(&self) -> Result<()> {
        let directive = self.filter.as_filter_directive();
        let env_filter = EnvFilter::try_new(&directive)?;
        self.raw.modify(|f| *f = env_filter)?;
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
    style: StyleConfig,
) -> (
    tracing_subscriber::reload::Layer<EnvFilter, InnerSubscriber>,
    ReloadHandle,
) {
    let (layer, raw_handle) =
        tracing_subscriber::reload::Layer::new(EnvFilter::new(level.as_filter_directive()));
    (
        layer,
        ReloadHandle {
            raw: raw_handle,
            filter: LogFilter::new(level.clone()),
            style,
        },
    )
}
