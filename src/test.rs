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
    let _layer = build_console_layer_with(&cfg, formatter);
}

#[cfg(feature = "file")]
#[test]
fn build_file_layer_creates_dirs() {
    let dir = std::env::temp_dir().join("caelum-test-filelayer");
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
    let dir = std::env::temp_dir().join("caelum-test-resolve");
    drop(std::fs::remove_dir_all(&dir));
    drop(std::fs::create_dir_all(&dir));
    let path = dir.join("new.log");

    let resolved = resolve_log_path(&path);
    assert_eq!(resolved, path);

    drop(std::fs::remove_dir_all(&dir));
}
