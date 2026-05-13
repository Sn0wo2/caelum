use tracing_subscriber::EnvFilter;
use tracing_subscriber::layer::Layered;

use crate::Result;
use crate::config::StyleConfig;
use crate::config::{Filter, Level};

pub(crate) type FmtLayer =
    Box<dyn tracing_subscriber::Layer<tracing_subscriber::Registry> + Send + Sync>;
pub(crate) type InnerSubscriber = Layered<FmtLayer, tracing_subscriber::Registry>;
type RawReloadHandle = tracing_subscriber::reload::Handle<EnvFilter, InnerSubscriber>;

#[must_use = "dropping ReloadHandle loses the ability to change log filters at runtime"]
pub struct ReloadHandle {
    raw: RawReloadHandle,
    filter: Filter,
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
        self.filter = Filter::new(Level::Custom(directive.to_owned()));
        Ok(())
    }

    pub fn set_filter(&mut self, filter: Filter) -> Result<()> {
        let directive = filter.as_directive();
        self.apply_directive(&directive)?;
        self.filter = filter;
        Ok(())
    }

    pub fn set_level(&mut self, level: Level) -> Result<()> {
        self.filter.level = level;
        self.apply_current_filter()
    }

    pub fn set_target_level(&mut self, target: impl Into<String>, level: Level) -> Result<()> {
        let target = target.into();
        self.filter.set_target(target, level);
        self.apply_current_filter()
    }

    pub fn remove_target_level(&mut self, target: &str) -> Result<()> {
        self.filter.remove_target(target);
        self.apply_current_filter()
    }

    fn apply_current_filter(&self) -> Result<()> {
        let directive = self.filter.as_directive();
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
    level: Level,
    style: StyleConfig,
) -> (
    tracing_subscriber::reload::Layer<EnvFilter, InnerSubscriber>,
    ReloadHandle,
) {
    let filter = Filter::new(level);
    let (layer, raw) = tracing_subscriber::reload::Layer::new(
        EnvFilter::try_new(filter.as_directive()).unwrap_or_default(),
    );

    (layer, ReloadHandle { raw, filter, style })
}
