use std::io;

use super::*;
use crate::reload::build_reload_filter_for_test as build_reload_filter;

#[test]
fn build_console_layer_all_variants() {
    let formats = [Format::Pretty, Format::Compact, Format::Json];
    let writers = [Writer::Stdout, Writer::Stderr];

    for format in &formats {
        for writer in &writers {
            let cfg = Console {
                format: format.clone(),
                writer: writer.clone(),
                ansi: true,
                show_path: true,
                show_spans: true,
                time_format: None,
                style: Style::default(),
            };
            let _layer = build_console_layer(&cfg);
        }
    }
}

#[test]
fn build_console_layer_no_ansi() {
    let cfg = Console {
        ansi: false,
        ..Default::default()
    };
    let _layer = build_console_layer(&cfg);
}

#[test]
fn build_console_layer_custom_time() {
    let cfg = Console {
        time_format: Some(String::from("%Y/%m/%d")),
        ..Default::default()
    };
    let _layer = build_console_layer(&cfg);
}

#[cfg(feature = "nerd")]
#[test]
fn build_console_layer_with_nerd_icons() {
    let cfg = Console::default();
    let formatter = Formatter::new().with_icons(Icons::nerd());
    let _layer = build_console_layer_with(&cfg, &formatter);
}

#[cfg(feature = "file")]
#[test]
fn build_file_layer_creates_dirs() {
    let dir = std::env::temp_dir().join("acta-test-filelayer");
    drop(std::fs::remove_dir_all(&dir));
    let nested = dir.join("a").join("b");
    let log_path = nested.join("app.log");

    let result = build_file_layer(&File {
        path: log_path,
        rotation: Rotation::None,
    });
    assert!(result.is_ok());

    let (_writer, _guard, path) = result.unwrap();
    assert!(path.parent().unwrap().exists());

    drop(std::fs::remove_dir_all(&dir));
}

#[test]
fn build_reload_filter_works() {
    let (_layer, mut handle, _subscriber) = build_reload_filter(Level::Info, Style::default());
    let result = handle.set_level(Level::Debug);
    assert!(result.is_ok(), "set_level failed: {:?}", result);
    assert!(handle.set_target_level("my_crate", Level::Trace).is_ok());
    assert!(handle.remove_target_level("my_crate").is_ok());
    assert!(handle.reload("info,my_crate=trace").is_ok());
    assert!(
        handle
            .set_filter(Filter::new(Level::Warn).with_target("my_crate", Level::Debug))
            .is_ok()
    );
}

#[cfg(feature = "file")]
#[test]
fn resolve_log_path_new_file() {
    use crate::writer::file::resolve_log_path;
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
    use crate::writer::file::resolve_log_path;
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
fn reload_handle_with_style_config() {
    let style = Style::default();
    let (_layer, mut handle, _subscriber) = build_reload_filter(Level::Info, style);
    handle.with_style(|s| s.theme = Theme::dracula());
    handle.with_style(|s| s.icons = Icons::unicode());
    handle.with_style(|s| s.labels = LevelLabels::SHORT);
}

#[test]
fn reload_handle_set_target_level_accepts_string() {
    let (_layer, mut handle, _subscriber) = build_reload_filter(Level::Info, Style::default());
    let target = String::from("my_crate");
    assert!(handle.set_target_level(target, Level::Trace).is_ok());
}

#[test]
fn reload_handle_remove_nonexistent_target_level() {
    let (_layer, mut handle, _subscriber) = build_reload_filter(Level::Info, Style::default());
    assert!(handle.remove_target_level("nonexistent_crate").is_ok());
}

#[test]
fn acta_error_display_lock_poisoned() {
    let msg = format!("{}", ActaError::LockPoisoned);
    assert!(msg.contains("log filter state lock poisoned"));
}

#[test]
fn acta_error_display_io() {
    let inner = io::Error::new(io::ErrorKind::NotFound, "test error");
    let msg = format!("{}", ActaError::Io(inner));
    assert!(msg.contains("I/O error"));
}

#[test]
fn acta_error_from_io_error() {
    let io_err = io::Error::new(io::ErrorKind::PermissionDenied, "access denied");
    let error: ActaError = io_err.into();
    assert!(matches!(error, ActaError::Io(_)));
}
