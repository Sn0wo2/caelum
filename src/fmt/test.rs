use crate::rotate_log_file;
use crate::config::LogRotation;
use super::*;

#[test]
fn ansi_formatter_defaults() {
    let fmt = AnsiFormatter::new();
    assert_eq!(fmt.time_format, "%H:%M:%S");
    assert_eq!(fmt.path_width, BUILD_PATH_WIDTH);
    assert!(fmt.show_path);
    assert!(fmt.show_spans);
}

#[test]
fn ansi_formatter_builder() {
    let fmt = AnsiFormatter::new()
        .with_time_format("%Y-%m-%d %H:%M:%S")
        .with_path_width(40)
        .with_show_path(false)
        .with_show_spans(false)
        .with_theme(Theme::monokai());

    assert_eq!(fmt.time_format, "%Y-%m-%d %H:%M:%S");
    assert_eq!(fmt.path_width, 40);
    assert!(!fmt.show_path);
    assert!(!fmt.show_spans);
}

#[cfg(feature = "nerd")]
#[test]
fn ansi_formatter_with_icons() {
    let fmt = AnsiFormatter::new().with_icons(Icons::nerd());
    assert_eq!(fmt.icons.bracket_open, "\u{e0b6}");
}

#[test]
fn theme_presets_are_distinct() {
    let s1 = format!("{:?}", Theme::trans_flag());
    let s2 = format!("{:?}", Theme::monokai());
    let s3 = format!("{:?}", Theme::dracula());
    assert_ne!(s1, s2);
    assert_ne!(s2, s3);
}

#[test]
fn theme_default_is_trans_flag() {
    assert_eq!(
        format!("{:?}", Theme::default()),
        format!("{:?}", Theme::trans_flag())
    );
}

#[cfg(feature = "nerd")]
#[test]
fn icons_unicode_vs_nerd() {
    let u = Icons::unicode();
    let n = Icons::nerd();
    assert_ne!(u.bracket_open, n.bracket_open);
    assert_ne!(u.bracket_close, n.bracket_close);
    assert_ne!(u.arrow, n.arrow);
    // separator is the same in both modes
    assert_eq!(u.separator, n.separator);
}

#[test]
fn smart_truncate_short_path() {
    let result = AnsiFormatter::smart_truncate("foo.rs", 10, 20);
    assert_eq!(result.len(), 20);
    assert!(result.contains("foo.rs:10"));
}

#[test]
fn smart_truncate_exact_width() {
    let result = AnsiFormatter::smart_truncate("foo.rs", 1, 8);
    assert_eq!(result, "foo.rs:1");
}

#[test]
fn smart_truncate_overflow() {
    let result = AnsiFormatter::smart_truncate("very/long/path/file.rs", 999, 15);
    assert!(result.len() <= 15);
}

#[test]
fn format_path_strips_src() {
    let result = AnsiFormatter::format_path("C:\\project\\src\\lib.rs", 42, 20);
    assert!(result.contains("lib.rs:42"));
    assert!(!result.contains("src/"));
}

#[test]
fn rotate_nonexistent_file_is_noop() {
    let path = std::env::temp_dir().join("sage-trace-test-nonexistent.log");
    let _ = std::fs::remove_file(&path);
    assert!(rotate_log_file(&path, LogRotation::Rename).is_ok());
    #[cfg(feature = "compress")]
    assert!(rotate_log_file(&path, LogRotation::Compress).is_ok());
    assert!(rotate_log_file(&path, LogRotation::None).is_ok());
}

#[test]
fn rotate_none_keeps_file() {
    let dir = std::env::temp_dir().join("sage-trace-test-fmt-none");
    let _ = std::fs::create_dir_all(&dir);
    let path = dir.join("app.log");
    std::fs::write(&path, b"hello\n").unwrap();

    rotate_log_file(&path, LogRotation::None).unwrap();
    assert!(path.exists());
    assert_eq!(std::fs::read_to_string(&path).unwrap(), "hello\n");
    let _ = std::fs::remove_dir_all(&dir);
}

#[test]
fn rotate_rename() {
    let dir = std::env::temp_dir().join("sage-trace-test-fmt-rename");
    let _ = std::fs::remove_dir_all(&dir);
    let _ = std::fs::create_dir_all(&dir);
    let path = dir.join("app.log");
    std::fs::write(&path, b"old content\n").unwrap();

    rotate_log_file(&path, LogRotation::Rename).unwrap();
    assert!(!path.exists());

    let entries: Vec<_> = std::fs::read_dir(&dir).unwrap().flatten().collect();
    assert_eq!(entries.len(), 1);
    let content = std::fs::read_to_string(entries[0].path()).unwrap();
    assert_eq!(content, "old content\n");
    let _ = std::fs::remove_dir_all(&dir);
}

#[test]
#[cfg(feature = "compress")]
fn rotate_compress() {
    let dir = std::env::temp_dir().join("sage-trace-test-fmt-compress");
    let _ = std::fs::remove_dir_all(&dir);
    let _ = std::fs::create_dir_all(&dir);
    let path = dir.join("app.log");
    std::fs::write(&path, b"compress me\n").unwrap();

    rotate_log_file(&path, LogRotation::Compress).unwrap();
    assert!(!path.exists());

    let entries: Vec<_> = std::fs::read_dir(&dir).unwrap().flatten().collect();
    assert_eq!(entries.len(), 1);
    let gz_data = std::fs::read(entries[0].path()).unwrap();
    assert!(gz_data.len() > 2);
    assert_eq!(gz_data[0], 0x1f);
    assert_eq!(gz_data[1], 0x8b);
    let _ = std::fs::remove_dir_all(&dir);
}
