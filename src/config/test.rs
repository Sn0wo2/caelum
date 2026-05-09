use super::*;

#[test]
fn logging_config_default() {
    let config = LoggingConfig::default();
    assert!(matches!(config.level, LogLevel::Info));
    assert!(config.console.is_some());
    assert!(config.file.is_none());

    let console = config.console.unwrap();
    assert!(matches!(console.format, LogFormat::Compact));
    assert!(console.ansi);
    assert!(matches!(console.writer, ConsoleWriter::Stdout));
    assert!(console.show_path);
    assert!(console.show_spans);
    assert!(console.time_format.is_none());
}

#[test]
fn console_config_default() {
    let cfg = ConsoleConfig::default();
    assert!(matches!(cfg.format, LogFormat::Compact));
    assert!(cfg.ansi);
    assert!(cfg.show_path);
    assert!(cfg.show_spans);
}

#[test]
fn log_level_directives() {
    assert_eq!(LogLevel::Error.as_filter_directive(), "error");
    assert_eq!(LogLevel::Warn.as_filter_directive(), "warn");
    assert_eq!(LogLevel::Info.as_filter_directive(), "info");
    assert_eq!(LogLevel::Debug.as_filter_directive(), "debug");
    assert_eq!(LogLevel::Trace.as_filter_directive(), "trace");
    assert_eq!(LogLevel::Off.as_filter_directive(), "off");
}

#[test]
fn log_level_custom_directive() {
    let level = LogLevel::Custom(FilterDirective::new("info,my_crate=debug"));
    assert_eq!(level.as_filter_directive(), "info,my_crate=debug");
}

#[test]
fn log_filter_builds_directive() {
    let filter = LogFilter::new(LogLevel::Debug)
        .with_target_level("my_crate", LogLevel::Trace)
        .with_target_level("my_crate::db", LogLevel::Warn);

    let directive = filter.as_filter_directive();

    assert!(directive.starts_with("debug,"));
    assert!(directive.contains("my_crate=trace"));
    assert!(directive.contains("my_crate::db=warn"));

    assert_eq!(directive.matches(',').count(), 2);
}

#[test]
fn log_filter_updates_targets() {
    let mut filter = LogFilter::new(LogLevel::Info);
    filter.set_target_level("my_crate", LogLevel::Debug);
    filter.set_target_level("my_crate", LogLevel::Trace);

    assert_eq!(filter.targets.len(), 1);
    assert_eq!(filter.as_filter_directive(), "info,my_crate=trace");
    assert!(filter.remove_target_level("my_crate"));
    assert_eq!(filter.as_filter_directive(), "info");
}

#[test]
fn log_rotation_default_is_none() {
    assert!(matches!(LogRotation::default(), LogRotation::None));
}
