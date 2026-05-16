use super::visitor::EventVisitor;
use super::*;
use crate::config::Rotation;
#[cfg(feature = "file")]
use crate::writer::file::rotate_log_file;
use smallvec::SmallVec;

#[test]
fn formatter_defaults() {
    let fmt = Formatter::new();
    assert_eq!(fmt.time_format, "%H:%M:%S");
    assert_eq!(fmt.path_width, BUILD_PATH_WIDTH);
    assert!(fmt.show_path);
    assert!(fmt.show_spans);
}

#[test]
fn formatter_builder() {
    let fmt = Formatter::new()
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
fn formatter_with_icons() {
    let fmt = Formatter::new().with_icons(Icons::nerd());
    assert_eq!(fmt.style_config().icons.bracket_open, "\u{e0b6}");
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
    assert_eq!(u.separator, n.separator);
}

#[test]
fn smart_truncate_short_path() {
    let fmt = Formatter::new().with_path_width(20);
    let result = fmt.format_path("foo.rs", 10);
    assert_eq!(result.len(), 20);
    assert!(result.contains("foo.rs:10"));
}

#[test]
fn smart_truncate_exact_width() {
    let fmt = Formatter::new().with_path_width(8);
    let result = fmt.format_path("foo.rs", 1);
    assert_eq!(result.as_str(), "foo.rs:1");
}

#[test]
fn smart_truncate_overflow() {
    let fmt = Formatter::new().with_path_width(15);
    let result = fmt.format_path("very/long/path/file.rs", 999);
    assert!(result.len() <= 15);
}

#[test]
fn smart_truncate_trailing_slash_before_filename() {
    let fmt = Formatter::new().with_path_width(20);
    let result = fmt.format_path("dir/subdir/file.rs", 42);
    assert!(result.len() <= 20);
    assert!(result.contains("file.rs:42"));
}

#[test]
fn smart_truncate_file_part_too_long() {
    let fmt = Formatter::new().with_path_width(15);
    let result = fmt.format_path("very_very_long_filename_test.rs", 10);
    assert!(
        result.len() <= 18,
        "result='{}', len={}",
        result,
        result.len()
    );
    assert!(result.starts_with('\u{2026}'));
}

#[test]
fn format_path_strips_src() {
    let fmt = Formatter::new().with_path_width(20);
    let result = fmt.format_path("C:\\project\\src\\lib.rs", 42);
    assert!(result.contains("lib.rs:42"));
    assert!(!result.contains("src/"));
}

#[cfg(feature = "file")]
#[test]
fn rotate_nonexistent_file_is_noop() {
    let path = std::env::temp_dir().join("acta-test-nonexistent.log");
    drop(std::fs::remove_file(&path));
    assert!(rotate_log_file(&path, Rotation::Rename).is_ok());
    #[cfg(feature = "compress")]
    assert!(rotate_log_file(&path, Rotation::Compress).is_ok());
    assert!(rotate_log_file(&path, Rotation::None).is_ok());
}

#[cfg(feature = "file")]
#[test]
fn rotate_none_keeps_file() {
    let dir = std::env::temp_dir().join("acta-test-fmt-none");
    drop(std::fs::create_dir_all(&dir));
    let path = dir.join("app.log");
    std::fs::write(&path, b"hello\n").unwrap();

    rotate_log_file(&path, Rotation::None).unwrap();
    assert!(path.exists());
    assert_eq!(std::fs::read_to_string(&path).unwrap(), "hello\n");
    drop(std::fs::remove_dir_all(&dir));
}

#[cfg(feature = "file")]
#[test]
fn rotate_rename() {
    let dir = std::env::temp_dir().join("acta-test-fmt-rename");
    drop(std::fs::remove_dir_all(&dir));
    drop(std::fs::create_dir_all(&dir));
    let path = dir.join("app.log");
    std::fs::write(&path, b"old content\n").unwrap();

    rotate_log_file(&path, Rotation::Rename).unwrap();
    assert!(!path.exists());

    let entries: Vec<_> = std::fs::read_dir(&dir).unwrap().flatten().collect();
    assert_eq!(entries.len(), 1);
    let content = std::fs::read_to_string(entries[0].path()).unwrap();
    assert_eq!(content, "old content\n");
    drop(std::fs::remove_dir_all(&dir));
}

#[test]
#[cfg(feature = "compress")]
fn rotate_compress() {
    let dir = std::env::temp_dir().join("acta-test-fmt-compress");
    drop(std::fs::remove_dir_all(&dir));
    drop(std::fs::create_dir_all(&dir));
    let path = dir.join("app.log");
    std::fs::write(&path, b"compress me\n").unwrap();

    rotate_log_file(&path, Rotation::Compress).unwrap();
    assert!(!path.exists());

    let entries: Vec<_> = std::fs::read_dir(&dir).unwrap().flatten().collect();
    assert_eq!(entries.len(), 1);
    let gz_data = std::fs::read(entries[0].path()).unwrap();
    assert!(gz_data.len() > 2);
    assert_eq!(gz_data[0], 0x1f);
    assert_eq!(gz_data[1], 0x8b);
    drop(std::fs::remove_dir_all(&dir));
}

#[test]
fn formatter_style_config_returns_reference() {
    let fmt = Formatter::new();
    let config = fmt.style_config();
    assert_eq!(config.labels.error, "E");
}

#[test]
fn formatter_with_style_config_replaces_all() {
    let config = Style {
        labels: LevelLabels::short(),
        icons: Icons::unicode(),
        theme: Theme::monokai(),
    };
    let fmt = Formatter::new().with_style_config(config);
    assert_eq!(fmt.style_config().labels, LevelLabels::short());
    assert_eq!(fmt.style_config().icons.bracket_open, "[");
    assert!((248..=255).contains(&fmt.style_config().theme.error.0));
}

#[test]
fn formatter_with_labels_changes_labels() {
    let fmt = Formatter::new();
    let before = fmt.style_config().labels.error;

    let fmt = fmt.with_labels(LevelLabels::long());
    assert_ne!(fmt.style_config().labels.error, before);
    assert_eq!(fmt.style_config().labels.error, "ERROR");
}

#[test]
fn formatter_with_icons_changes_icons() {
    let fmt = Formatter::new().with_icons(Icons::unicode());
    assert_eq!(fmt.style_config().icons.bracket_open, "[");
}

#[test]
fn formatter_with_theme_changes_theme() {
    let fmt = Formatter::new().with_theme(Theme::monokai());
    assert_ne!(
        format!("{:?}", fmt.style_config().theme),
        format!("{:?}", Theme::default())
    );
}

#[test]
fn event_visitor_records_message_field() {
    let mut visitor = EventVisitor::default();
    visitor.record_field("message", "hello".to_owned());
    assert_eq!(visitor.message, Some("hello".to_owned()));
    assert!(visitor.fields.is_empty());
}

#[test]
fn event_visitor_records_msg_alias() {
    let mut visitor = EventVisitor::default();
    visitor.record_field("msg", "world".to_owned());
    assert_eq!(visitor.message, Some("world".to_owned()));
    assert!(visitor.fields.is_empty());
}

#[test]
fn event_visitor_records_other_fields_as_pairs() {
    let mut visitor = EventVisitor::default();
    visitor.record_field("user", "alice".to_owned());
    visitor.record_field("count", "42".to_owned());
    assert!(visitor.message.is_none());
    assert_eq!(
        visitor.fields,
        SmallVec::<[(String, String); 4]>::from_vec(vec![
            ("user".to_owned(), "alice".to_owned()),
            ("count".to_owned(), "42".to_owned())
        ])
    );
}

#[test]
fn event_visitor_default_has_no_message_and_empty_fields() {
    let visitor = EventVisitor::default();
    assert!(visitor.message.is_none());
    assert!(visitor.fields.is_empty());
}

#[test]
fn event_visitor_order_preserved_message_extracted() {
    let mut visitor = EventVisitor::default();
    visitor.record_field("x", "1".to_owned());
    visitor.record_field("message", "the message".to_owned());
    visitor.record_field("y", "2".to_owned());
    assert_eq!(visitor.message, Some("the message".to_owned()));
    assert_eq!(
        visitor.fields,
        SmallVec::<[(String, String); 4]>::from_vec(vec![
            ("x".to_owned(), "1".to_owned()),
            ("y".to_owned(), "2".to_owned())
        ])
    );
}

#[test]
fn level_labels_short() {
    let labels = LevelLabels::short();
    assert_eq!(labels.error, "E");
    assert_eq!(labels.warn, "W");
    assert_eq!(labels.info, "I");
    assert_eq!(labels.debug, "D");
    assert_eq!(labels.trace, "T");
}

#[test]
fn level_labels_long() {
    let labels = LevelLabels::long();
    assert_eq!(labels.error, "ERROR");
    assert_eq!(labels.warn, " WARN");
    assert_eq!(labels.info, " INFO");
    assert_eq!(labels.debug, "DEBUG");
    assert_eq!(labels.trace, "TRACE");
}

#[test]
fn level_labels_default_equals_short() {
    assert_eq!(LevelLabels::default(), LevelLabels::short());
}

#[test]
fn icons_unicode() {
    let icons = Icons::unicode();
    assert_eq!(icons.bracket_open, "[");
    assert_eq!(icons.bracket_close, "]");
    assert_eq!(icons.separator, "\u{2507}");
    assert_eq!(icons.arrow, ">");
    assert_eq!(icons.span_delimiter, "->");
}

#[test]
fn icons_is_nerd_returns_false_for_unicode() {
    let icons = Icons::unicode();
    assert!(!icons.is_nerd());
}

#[cfg(feature = "nerd")]
#[test]
fn icons_is_nerd_returns_true_for_nerd() {
    let icons = Icons::nerd();
    assert!(icons.is_nerd());
}

#[test]
fn style_config_default() {
    let config = Style::default();
    assert_eq!(
        format!("{:?}", config.theme),
        format!("{:?}", Theme::trans_flag())
    );
    #[cfg(feature = "nerd")]
    assert_eq!(config.icons, Icons::nerd());
    #[cfg(not(feature = "nerd"))]
    assert_eq!(config.icons, Icons::unicode());
    assert_eq!(config.labels, LevelLabels::short());
}

#[test]
fn theme_all_have_distinct_accent_colors() {
    let themes = [
        Theme::trans_flag(),
        Theme::monokai(),
        Theme::dracula(),
        Theme::nord(),
        Theme::catppuccin_mocha(),
        Theme::gruvbox(),
        Theme::one_dark(),
        Theme::tokyo_night(),
    ];
    for i in 0..themes.len() {
        for j in (i + 1)..themes.len() {
            assert_ne!(format!("{:?}", themes[i].accent), format!("{:?}", themes[j].accent));
        }
    }
}

#[test]
fn theme_default_equals_trans_flag() {
    assert_eq!(
        format!("{:?}", Theme::default()),
        format!("{:?}", Theme::trans_flag())
    );
}
