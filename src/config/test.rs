use super::*;

#[test]
fn config_default() {
    let config = Config::default();
    assert!(matches!(config.level, Level::Info));
    assert!(config.console.is_some());
    assert!(config.file.is_none());

    let console = config.console.unwrap();
    assert!(matches!(console.format, Format::Compact));
    assert!(console.ansi);
    assert!(matches!(console.writer, Writer::Stdout));
    assert!(console.show_path);
    assert!(console.show_spans);
    assert!(console.time_format.is_none());
}

#[test]
fn console_default() {
    let cfg = Console::default();
    assert!(matches!(cfg.format, Format::Compact));
    assert!(cfg.ansi);
    assert!(cfg.show_path);
    assert!(cfg.show_spans);
}

#[test]
fn console_builder() {
    let cfg = Console::builder()
        .format(Format::Json)
        .ansi(false)
        .show_path(false)
        .show_spans(false)
        .time_format("%Y")
        .style(Style::default())
        .build();
    assert!(matches!(cfg.format, Format::Json));
    assert!(!cfg.ansi);
    assert!(!cfg.show_path);
    assert!(!cfg.show_spans);
    assert_eq!(cfg.time_format, Some("%Y".to_string()));
}

#[test]
fn config_builder() {
    let cfg = Config::builder()
        .level(Level::Debug)
        .console(Console::new())
        .build();
    assert_eq!(cfg.level.as_directive(), "debug");
    assert!(cfg.console.is_some());
}

#[test]
fn level_directives() {
    assert_eq!(Level::Error.as_directive(), "error");
    assert_eq!(Level::Warn.as_directive(), "warn");
    assert_eq!(Level::Info.as_directive(), "info");
    assert_eq!(Level::Debug.as_directive(), "debug");
    assert_eq!(Level::Trace.as_directive(), "trace");
    assert_eq!(Level::Off.as_directive(), "off");
}

#[test]
fn level_custom_directive() {
    let level = Level::Custom("info,my_crate=debug".into());
    assert_eq!(level.as_directive(), "info,my_crate=debug");
}

#[test]
fn filter_builds_directive() {
    let filter = Filter::new(Level::Debug)
        .with_target("my_crate", Level::Trace)
        .with_target("my_crate::db", Level::Warn);

    let directive = filter.as_directive();
    assert!(directive.starts_with("debug,"));
    assert!(directive.contains("my_crate=trace"));
    assert!(directive.contains("my_crate::db=warn"));
    assert_eq!(directive.matches(',').count(), 2);
}

#[test]
fn filter_updates_targets() {
    let mut filter = Filter::new(Level::Info);
    filter.set_target("my_crate", Level::Debug);
    filter.set_target("my_crate", Level::Trace);

    assert_eq!(filter.as_directive(), "info,my_crate=trace");
    assert!(filter.remove_target("my_crate"));
    assert_eq!(filter.as_directive(), "info");
}

#[test]
fn rotation_default_is_none() {
    assert!(matches!(Rotation::default(), Rotation::None));
}

#[test]
fn filter_remove_target_exists() {
    let mut filter = Filter::new(Level::Info);
    filter.set_target("my_crate", Level::Debug);
    assert!(filter.remove_target("my_crate"));
}

#[test]
fn filter_remove_target_not_exists() {
    let mut filter = Filter::new(Level::Info);
    assert!(!filter.remove_target("nonexistent"));
}

#[test]
fn filter_from_level() {
    let filter: Filter = Level::Warn.into();
    assert_eq!(filter.as_directive(), "warn");
}

#[test]
fn filter_from_level_debug() {
    let filter: Filter = Level::Debug.into();
    assert_eq!(filter.as_directive(), "debug");
}
