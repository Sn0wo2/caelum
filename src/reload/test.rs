use tracing_subscriber::layer::Layered;
use tracing_subscriber::prelude::*;

use crate::config::{Filter, Level, Style};
use crate::reload::{FmtLayer, InnerSubscriber, ReloadHandle};

pub type SubscriberWithBoth = Layered<
    tracing_subscriber::reload::Layer<tracing_subscriber::EnvFilter, InnerSubscriber>,
    InnerSubscriber,
>;

pub fn build_reload_filter_for_test(
    level: Level,
    style: Style,
) -> (
    tracing_subscriber::reload::Layer<FmtLayer, tracing_subscriber::Registry>,
    ReloadHandle,
    SubscriberWithBoth,
) {
    let console_layer = tracing_subscriber::fmt::Layer::default()
        .with_writer(std::io::sink)
        .boxed();
    let subscriber_with_console = tracing_subscriber::Registry::default().with(console_layer);

    let filter = Filter::new(level);
    let env_filter =
        tracing_subscriber::EnvFilter::try_new(filter.as_directive()).unwrap_or_default();
    let (env_layer, env_handle) = tracing_subscriber::reload::Layer::new(env_filter);
    let subscriber = subscriber_with_console.with(env_layer);

    let reload_handle = ReloadHandle {
        raw: env_handle,
        filter,
        style,
    };

    let console_layer = tracing_subscriber::fmt::Layer::default()
        .with_writer(std::io::sink)
        .boxed();
    let (layer, _) = tracing_subscriber::reload::Layer::new(console_layer);
    (layer, reload_handle, subscriber)
}
