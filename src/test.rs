use super::*;

#[test]
fn build_console_layer_all_variants() {
    let formats = [LogFormat::Pretty, LogFormat::Compact, LogFormat::Json];
    let writers = [ConsoleWriter::Stdout, ConsoleWriter::Stderr];

    for format in &formats {
        for writer in &writers {
            let cfg = ConsoleConfig {
                format: format.clone(),
                writer: writer.clone(),
                ansi: true,
                show_path: true,
                show_spans: true,
                time_format: None,
            };
            let _layer = build_console_layer(&cfg);
        }
    }
}

#[test]
fn build_console_layer_no_ansi() {
    let cfg = ConsoleConfig {
        ansi: false,
        ..Default::default()
    };
    let _layer = build_console_layer(&cfg);
}

#[test]
fn build_console_layer_custom_time() {
    let cfg = ConsoleConfig {
        time_format: Some(String::from("%Y/%m/%d")),
        ..Default::default()
    };
    let _layer = build_console_layer(&cfg);
}

#[cfg(feature = "nerd")]
#[test]
fn build_console_layer_with_nerd_icons() {
    let cfg = ConsoleConfig::default();
    let formatter = AnsiFormatter::new().with_icons(Icons::nerd());
    let _layer = build_console_layer_with(&cfg, &formatter);
}

#[cfg(feature = "file")]
#[test]
fn build_file_layer_creates_dirs() {
    let dir = std::env::temp_dir().join("acta-test-filelayer");
    drop(std::fs::remove_dir_all(&dir));
    let nested = dir.join("a").join("b");
    let log_path = nested.join("app.log");

    let result = build_file_layer(&FileLoggingConfig {
        path: log_path,
        rotation: LogRotation::None,
    });
    assert!(result.is_ok());

    let r = result.unwrap();
    assert!(r.path.parent().unwrap().exists());
    drop(r.guard);

    drop(std::fs::remove_dir_all(&dir));
}

#[test]
fn build_reload_filter_works() {
    let (_layer, handle) = build_reload_filter(&LogLevel::Info, None);
    assert!(handle.set_level(LogLevel::Debug).is_ok());
    assert!(handle.set_target_level("my_crate", LogLevel::Trace).is_ok());
    assert!(handle.remove_target_level("my_crate").is_ok());
    assert!(handle.reload("info,my_crate=trace").is_ok());
    assert!(
        handle
            .set_filter(
                LogFilter::new(LogLevel::Warn).with_target_level("my_crate", LogLevel::Debug)
            )
            .is_ok()
    );
}

#[cfg(feature = "file")]
#[test]
fn resolve_log_path_new_file() {
    let dir = std::env::temp_dir().join("acta-test-resolve");
    drop(std::fs::remove_dir_all(&dir));
    drop(std::fs::create_dir_all(&dir));
    let path = dir.join("new.log");

    let resolved = resolve_log_path(&path);
    assert_eq!(resolved, path);

    drop(std::fs::remove_dir_all(&dir));
}

#[cfg(feature = "file")]
#[test]
fn resolve_log_path_fallback_when_parent_is_file() {
    let dir = std::env::temp_dir().join("acta-test-resolve-fallback");
    drop(std::fs::remove_dir_all(&dir));
    drop(std::fs::create_dir_all(&dir));

    let bad_file = dir.join("existing_file");
    std::fs::write(&bad_file, b"contents").unwrap();

    let nested = bad_file.join("should_not_exist.log");
    let resolved = resolve_log_path(&nested);

    assert!(!resolved.exists());
    let pid = std::process::id();
    assert!(
        resolved
            .file_name()
            .unwrap()
            .to_str()
            .unwrap()
            .contains(&pid.to_string())
    );

    drop(std::fs::remove_dir_all(&dir));
}

#[test]
fn reload_handle_set_theme_with_style_config() {
    let style = StyleConfig::default();
    let (_layer, mut handle) = build_reload_filter(&LogLevel::Info, Some(style));
    assert!(handle.set_theme(Theme::dracula()).is_ok());
}

#[test]
fn reload_handle_set_icons_with_style_config() {
    let style = StyleConfig::default();
    let (_layer, mut handle) = build_reload_filter(&LogLevel::Info, Some(style));
    assert!(handle.set_icons(Icons::unicode()).is_ok());
}

#[test]
fn reload_handle_set_labels_with_style_config() {
    let style = StyleConfig::default();
    let (_layer, mut handle) = build_reload_filter(&LogLevel::Info, Some(style));
    assert!(handle.set_labels(LevelLabels::short()).is_ok());
}

#[test]
fn reload_handle_set_theme_without_style_config() {
    let (_layer, mut handle) = build_reload_filter(&LogLevel::Info, None);
    assert!(handle.set_theme(Theme::monokai()).is_err());
    assert!(handle.set_icons(Icons::unicode()).is_err());
    assert!(handle.set_labels(LevelLabels::long()).is_err());
}

#[test]
fn reload_handle_set_target_level_accepts_string() {
    let (_layer, handle) = build_reload_filter(&LogLevel::Info, None);
    let target = String::from("my_crate");
    assert!(handle.set_target_level(target, LogLevel::Trace).is_ok());
}

#[test]
fn reload_handle_remove_nonexistent_target_level() {
    let (_layer, handle) = build_reload_filter(&LogLevel::Info, None);
    assert!(handle.remove_target_level("nonexistent_crate").is_ok());
}
#[test]
fn acta_error_display_lock_poisoned() {
    let msg = format!("{}", ActaError::LockPoisoned);
    assert!(msg.contains("log filter state lock poisoned"));
}

#[test]
fn acta_error_display_style_not_configured() {
    let msg = format!("{}", ActaError::StyleNotConfigured);
    assert!(msg.contains("formatter style reload not configured"));
}

#[test]
fn acta_error_display_io() {
    let inner = std::io::Error::new(std::io::ErrorKind::NotFound, "test error");
    let msg = format!("{}", ActaError::Io(inner));
    assert!(msg.contains("I/O error"));
}

#[test]
fn acta_error_from_io_error() {
    let io_err = std::io::Error::new(std::io::ErrorKind::PermissionDenied, "access denied");
    let error: ActaError = io_err.into();
    assert!(matches!(error, ActaError::Io(_)));
}
