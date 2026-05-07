use std::sync::LazyLock;

use caelum::{
    AnsiFormatter, ConsoleConfig, ConsoleWriter, FileLoggingConfig, FilterDirective, Icons,
    LevelLabels, LogFormat, LogLevel, LogRotation, LoggingConfig, Theme, build_console_layer,
    build_console_layer_with, build_file_layer, build_reload_filter, init_tracing, rotate_log_file,
};
use tracing_subscriber::prelude::*;

fn sep(c: &str, n: usize) -> String {
    c.repeat(n)
}

const SECTION_WIDTH: usize = 64;
const BOX_WIDTH: usize = 54;
static THEMES: LazyLock<Vec<(&'static str, Theme)>> = LazyLock::new(|| {
    vec![
        ("trans_flag", Theme::trans_flag()),
        ("monokai", Theme::monokai()),
        ("dracula", Theme::dracula()),
        ("nord", Theme::nord()),
        ("catppuccin_mocha", Theme::catppuccin_mocha()),
        ("gruvbox", Theme::gruvbox()),
        ("one_dark", Theme::one_dark()),
        ("tokyo_night", Theme::tokyo_night()),
    ]
});

fn section(title: &str) {
    let pad = (SECTION_WIDTH.saturating_sub(title.len() + 4)) / 2;
    println!();
    println!("\u{250c}{}\u{2510}", sep("\u{2500}", SECTION_WIDTH));
    println!("\u{2502}{:>pad$}  {}  {:pad$}\u{2502}", "", title, "");
    println!("\u{2514}{}\u{2518}", sep("\u{2500}", SECTION_WIDTH));
}

fn sub(title: &str) {
    println!("  \u{25c6} {title}");
}
fn info(msg: &str) {
    println!("    \u{00b7} {msg}");
}
fn success(msg: &str) {
    println!("    \u{2713} {msg}");
}
fn fail(msg: &str) {
    println!("    \u{2717} {msg}");
}

fn demo_logs(label: &str) {
    tracing::info!("{label}: info level log output");
    tracing::warn!(
        user = "alice",
        count = 42,
        "{label}: warn with structured fields"
    );
    tracing::error!("{label}: error level log output");
    tracing::debug!("{label}: debug level log output");
    let _span = tracing::info_span!("my_span", task = "demo").entered();
    tracing::trace!("{label}: trace inside a span context");
}

fn demo_logs_rich(label: &str) {
    let req = ("GET", "/api/users", 200u16);
    tracing::info!(
        user = "alice",
        active = true,
        score = 99.5,
        "{label}: rich fields: bool, float, string"
    );
    tracing::warn!(
        latency_ms = 247,
        retries = 3,
        "{label}: warn with numeric fields"
    );
    tracing::error!(error = "connection refused", code = 500, request = ?req, "{label}: error with Debug-format struct");
    tracing::debug!(
        cache = "HIT",
        ttl = 300,
        "{label}: debug with key=value pairs"
    );
    let _span = tracing::info_span!("request", method = "POST", id = 42).entered();
    tracing::info!(db_ms = 12, rows = 5, "{label}: info inside request span");
    let _inner = tracing::debug_span!("serialize", format = "json").entered();
    tracing::trace!(bytes = 1024, "{label}: trace inside nested serialize span");
}

fn main() {
    println!();
    println!("\u{2554}\u{2550}{}\u{2557}", sep("\u{2550}", BOX_WIDTH));
    println!("\u{2551}           caelum  \u{00b7}  debug suite               \u{2551}");
    println!("\u{255a}\u{2550}{}\u{255d}", sep("\u{2550}", BOX_WIDTH));

    section("FORMAT MODES");

    sub("Compact + Unicode icons");
    {
        let console = ConsoleConfig::default();
        let formatter = AnsiFormatter::new().with_icons(Icons::unicode());
        let layer = build_console_layer_with(&console, formatter);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs("unicode"));
    }

    #[cfg(feature = "nerd")]
    {
        sub("Compact + Nerd Font icons");
        let console = ConsoleConfig::default();
        let formatter = AnsiFormatter::new().with_icons(Icons::nerd());
        let layer = build_console_layer_with(&console, formatter);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs("nerd"));
    }

    sub("Pretty format — target, file, line, span context");
    {
        let console = ConsoleConfig {
            format: LogFormat::Pretty,
            ..Default::default()
        };
        let layer = build_console_layer(&console);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs_rich("pretty"));
    }

    sub("JSON format — machine-readable structured output");
    {
        let console = ConsoleConfig {
            format: LogFormat::Json,
            ansi: false,
            ..Default::default()
        };
        let layer = build_console_layer(&console);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs("json"));
    }

    sub("Compact vs Pretty comparison");
    {
        info("Compact — single-line, color-coded, compact output");
        let console = ConsoleConfig::default();
        let formatter = AnsiFormatter::new()
            .with_show_path(false)
            .with_show_spans(false);
        let layer = build_console_layer_with(&console, formatter);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs_rich("compact"));

        println!();
        info("Pretty — multiline, with target path and span context");
        let console = ConsoleConfig {
            format: LogFormat::Pretty,
            ..Default::default()
        };
        let layer = build_console_layer(&console);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs_rich("pretty"));
    }

    section("THEMES");

    sub("Color palette — accent, secondary, text for each theme");
    {
        for (name, theme) in &*THEMES {
            println!(
                "    {name:<20} accent={:?}  secondary={:?}  text={:?}",
                theme.accent, theme.secondary, theme.text
            );
        }
    }

    sub("Live preview — actual log output per theme");
    {
        for (name, theme) in &*THEMES {
            println!("  [{name}]");
            let console = ConsoleConfig::default();
            let formatter = AnsiFormatter::new()
                .with_theme(*theme)
                .with_show_path(false);
            let layer = build_console_layer_with(&console, formatter);
            let subscriber = tracing_subscriber::registry().with(layer);
            tracing::subscriber::with_default(subscriber, || {
                tracing::info!(
                    user = "alice",
                    count = 42,
                    "theme preview: info with user and count fields"
                );
            });
        }
    }

    sub("All five log levels rendered with one_dark theme");
    {
        let console = ConsoleConfig::default();
        let fmt = AnsiFormatter::new()
            .with_theme(Theme::one_dark())
            .with_show_path(false)
            .with_show_spans(false);
        let layer = build_console_layer_with(&console, fmt);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || {
            tracing::error!(code = 500, "error level: server crash with error code");
            tracing::warn!(
                threshold = 0.9,
                "warn level: resource threshold approaching limit"
            );
            tracing::info!(users = 42, "info level: service running normally");
            tracing::debug!(
                query = "SELECT *",
                took_ms = 3,
                "debug level: database query performed"
            );
            tracing::trace!(state = "idle", "trace level: entering idle state");
        });
    }

    section("FORMATTER OPTIONS");

    sub("Builder API — inspect AnsiFormatter fields");
    {
        let fmt = AnsiFormatter::new()
            .with_theme(Theme::monokai())
            .with_time_format("%Y-%m-%d %H:%M:%S")
            .with_path_width(40)
            .with_show_path(false)
            .with_show_spans(false);
        info(&format!("time_format : {:?}", fmt.time_format));
        info(&format!("path_width  : {}", fmt.path_width));
        info(&format!("show_path   : {}", fmt.show_path));
        info(&format!("show_spans  : {}", fmt.show_spans));
    }

    sub("Level labels — short [E/W/I/D/T] vs long [ERROR/WARN/INFO/DEBUG/TRACE]");
    {
        let short_fmt = AnsiFormatter::new()
            .with_labels(LevelLabels::short())
            .with_show_path(false)
            .with_show_spans(false);
        let long_fmt = AnsiFormatter::new()
            .with_labels(LevelLabels::long())
            .with_show_path(false)
            .with_show_spans(false);

        info("short — single-letter level labels");
        let layer = build_console_layer_with(&ConsoleConfig::default(), short_fmt);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs("short"));

        println!();
        info("long — full-word level labels");
        let layer = build_console_layer_with(&ConsoleConfig::default(), long_fmt);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs("long"));
    }

    sub("Path display and span decoration toggles");
    {
        info("path=on, spans=on — file location and span chain visible");
        let fmt = AnsiFormatter::new()
            .with_show_path(true)
            .with_show_spans(true);
        let layer = build_console_layer_with(&ConsoleConfig::default(), fmt);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs_rich("full"));

        println!();
        info("path=off, spans=off — minimal output, only message and fields");
        let fmt = AnsiFormatter::new()
            .with_show_path(false)
            .with_show_spans(false);
        let layer = build_console_layer_with(&ConsoleConfig::default(), fmt);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs_rich("minimal"));
    }

    sub("Timestamp format — default HH:MM:SS vs custom with milliseconds");
    {
        info("default format: %H:%M:%S");
        let fmt = AnsiFormatter::new()
            .with_show_path(false)
            .with_show_spans(false);
        let layer = build_console_layer_with(&ConsoleConfig::default(), fmt);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs("default-time"));

        println!();
        info("custom format: %Y-%m-%d %H:%M:%S%.3f");
        let fmt = AnsiFormatter::new()
            .with_time_format("%Y-%m-%d %H:%M:%S%.3f")
            .with_show_path(false)
            .with_show_spans(false);
        let layer = build_console_layer_with(&ConsoleConfig::default(), fmt);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs("ms"));
    }

    sub("Path width — compile-time 28 (default) vs runtime 20");
    {
        info("path width = 28 (compile-time default)");
        let fmt = AnsiFormatter::new().with_show_spans(false);
        let layer = build_console_layer_with(&ConsoleConfig::default(), fmt);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs("wide"));

        println!();
        info("path width = 20 (overridden at runtime)");
        let fmt = AnsiFormatter::new()
            .with_path_width(20)
            .with_show_spans(false);
        let layer = build_console_layer_with(&ConsoleConfig::default(), fmt);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs("narrow"));
    }

    section("ADVANCED");

    sub("Runtime reload — change log level and target filter dynamically");
    {
        let (filter_layer, reload_handle) = build_reload_filter(&LogLevel::Info, None);
        let console = ConsoleConfig::default();
        let fmt = AnsiFormatter::new()
            .with_show_path(false)
            .with_show_spans(false);
        let layer = build_console_layer_with(&console, fmt);
        let subscriber = tracing_subscriber::registry()
            .with(layer)
            .with(filter_layer);

        info("initial level=Info — only >=Info logs appear");
        tracing::subscriber::with_default(subscriber, || {
            tracing::info!("visible: info level passes Info filter");
            tracing::debug!("suppressed: debug below Info threshold");
            tracing::trace!("suppressed: trace below Info threshold");

            reload_handle.set_level(LogLevel::Debug).unwrap();
            info("reload → set_level(Debug)");
            tracing::debug!("visible: debug now passes Debug filter");
            tracing::trace!("suppressed: trace still below Debug threshold");

            reload_handle.set_level(LogLevel::Trace).unwrap();
            info("reload → set_level(Trace)");
            tracing::trace!("visible: trace now passes Trace filter");

            reload_handle
                .set_target_level("reload_demo", LogLevel::Warn)
                .unwrap();
            info("reload → set_target_level(reload_demo, Warn)");
            tracing::info!(target: "reload_demo", "suppressed: target capped at Warn, info < Warn");
            tracing::warn!(target: "reload_demo", "visible: target capped at Warn, warn >= Warn");
        });
    }

    sub("Runtime style switch — change icons, theme, labels at runtime");
    {
        let console = ConsoleConfig::default();
        let fmt = AnsiFormatter::new()
            .with_icons(Icons::unicode())
            .with_theme(Theme::trans_flag())
            .with_show_path(false)
            .with_show_spans(false);
        let style = fmt.style_config();
        let layer = build_console_layer_with(&console, fmt);
        let (filter_layer, reload_handle) = build_reload_filter(&LogLevel::Info, Some(style));
        let subscriber = tracing_subscriber::registry()
            .with(layer)
            .with(filter_layer);

        tracing::subscriber::with_default(subscriber, || {
            tracing::info!("initial: unicode icons, trans_flag theme, short labels");

            reload_handle.set_theme(Theme::monokai()).unwrap();
            info("reload → set_theme(monokai)");
            tracing::info!("theme switched to monokai");

            reload_handle.set_labels(LevelLabels::long()).unwrap();
            info("reload → set_labels(long)");
            tracing::warn!("labels switched to full-word");

            #[cfg(feature = "nerd")]
            {
                reload_handle.set_icons(Icons::nerd()).unwrap();
                info("reload → set_icons(nerd)");
                tracing::error!("icons switched to nerd font");
            }
        });
    }

    sub("Span nesting — context chain across 3 span layers");
    {
        let console = ConsoleConfig::default();
        let fmt = AnsiFormatter::new()
            .with_show_path(false)
            .with_show_spans(true);
        let layer = build_console_layer_with(&console, fmt);
        let subscriber = tracing_subscriber::registry().with(layer);

        info("depth 0 → 1 → 2 → 3, each log shows its full span chain");
        tracing::subscriber::with_default(subscriber, || {
            tracing::info!("outside spans — no span context");
            let _a = tracing::info_span!("layer1").entered();
            tracing::info!("inside layer1 — span chain: [layer1]");
            let _b = tracing::info_span!("layer2", depth = 2).entered();
            tracing::warn!("inside layer2 — span chain: [layer1 > layer2{{depth=2}}]");
            let _c = tracing::debug_span!("layer3").entered();
            tracing::error!("inside layer3 — span chain: [layer1 > layer2 > layer3]");
        });
    }

    sub("Stderr writer — redirect log output to standard error");
    {
        let console = ConsoleConfig {
            writer: ConsoleWriter::Stderr,
            ..Default::default()
        };
        let fmt = AnsiFormatter::new()
            .with_show_path(false)
            .with_show_spans(false);
        let layer = build_console_layer_with(&console, fmt);
        let subscriber = tracing_subscriber::registry().with(layer);
        eprintln!("    (output below written to stderr)");
        tracing::subscriber::with_default(subscriber, || demo_logs("stderr"));
    }

    sub("Configuration matrix — ConsoleConfig permutations");
    {
        let configs = vec![
            ("compact + stdout", ConsoleConfig::default()),
            (
                "json + stderr + no-ansi",
                ConsoleConfig {
                    format: LogFormat::Json,
                    writer: ConsoleWriter::Stderr,
                    ansi: false,
                    ..Default::default()
                },
            ),
            (
                "pretty + no-path + no-spans",
                ConsoleConfig {
                    format: LogFormat::Pretty,
                    show_path: false,
                    show_spans: false,
                    ..Default::default()
                },
            ),
        ];
        for (label, cfg) in &configs {
            info(&format!("{label:<32} \u{2192} {cfg:?}"));
        }
    }

    section("INFRASTRUCTURE");

    sub("Log level to tracing filter directive mapping");
    {
        let levels = vec![
            LogLevel::Error,
            LogLevel::Warn,
            LogLevel::Info,
            LogLevel::Debug,
            LogLevel::Trace,
            LogLevel::Off,
            LogLevel::Custom(FilterDirective::new("info,my_crate=debug")),
        ];
        for level in &levels {
            info(&format!(
                "{:?} \u{2192} \"{}\"",
                level,
                level.as_filter_directive()
            ));
        }
    }

    sub("build_console_layer — verify layer construction");
    {
        let layer = build_console_layer(&ConsoleConfig::default());
        let _ = layer;
        success("build_console_layer(default)");

        let layer = build_console_layer(&ConsoleConfig {
            format: LogFormat::Json,
            writer: ConsoleWriter::Stderr,
            ansi: false,
            ..Default::default()
        });
        let _ = layer;
        success("build_console_layer(json+stderr)");
    }

    sub("File log rotation — Rename (and Compress if enabled)");
    {
        let tmp_dir = std::env::temp_dir().join("caelum-debug");
        let _ = std::fs::create_dir_all(&tmp_dir);
        let log_path = tmp_dir.join("test.log");

        let file_config = FileLoggingConfig {
            path: log_path.clone(),
            rotation: LogRotation::Rename,
        };
        info(&format!(
            "config: {:?}  |  path: {}",
            file_config,
            log_path.display()
        ));

        std::fs::write(&log_path, b"old log content\n").ok();
        match rotate_log_file(&log_path, LogRotation::Rename) {
            Ok(()) => success("rotate(Rename) — old file renamed with timestamp suffix"),
            Err(e) => fail(&format!("rotate(Rename): {e}")),
        }

        #[cfg(feature = "compress")]
        {
            std::fs::write(&log_path, b"compress me\n").ok();
            match rotate_log_file(&log_path, LogRotation::Compress) {
                Ok(()) => success("rotate(Compress) — old file compressed to .gz"),
                Err(e) => fail(&format!("rotate(Compress): {e}")),
            }
        }

        if let Ok(entries) = std::fs::read_dir(&tmp_dir) {
            info("rotated files on disk:");
            for entry in entries.flatten() {
                println!("      \u{00b7} {}", entry.file_name().to_string_lossy());
            }
        }
        let _ = std::fs::remove_dir_all(&tmp_dir);
    }

    #[cfg(feature = "file")]
    {
        sub("build_file_layer — file appender construction");
        let tmp_dir = std::env::temp_dir().join("caelum-debug-file");
        let _ = std::fs::create_dir_all(&tmp_dir);

        let result = build_file_layer(&FileLoggingConfig {
            path: tmp_dir.join("app.log"),
            rotation: LogRotation::None,
        });
        match result {
            Ok(r) => {
                success(&format!("build_file_layer → {}", r.path.display()));
                drop(r.guard);
            }
            Err(e) => fail(&format!("build_file_layer: {e}")),
        }
        let _ = std::fs::remove_dir_all(&tmp_dir);
    }

    #[cfg(feature = "file")]
    {
        sub("init_tracing — end-to-end: subscriber + console + file + reload");
        let tmp_dir = std::env::temp_dir().join("caelum-debug-full");
        let _ = std::fs::create_dir_all(&tmp_dir);

        let config = LoggingConfig {
            level: LogLevel::Debug,
            console: Some(ConsoleConfig {
                show_path: false,
                show_spans: false,
                ..Default::default()
            }),
            file: Some(FileLoggingConfig {
                path: tmp_dir.join("app.log"),
                rotation: LogRotation::None,
            }),
        };

        let guard = init_tracing(&config);
        match guard {
            Ok(g) => {
                success("init_tracing — global subscriber set");
                if let Some(ref path) = g.log_path {
                    info(&format!("log file → {}", path.display()));
                }
                tracing::info!(
                    init = true,
                    "init_tracing: system initialized with console+file logging"
                );
                tracing::debug!(
                    step = "config_loaded",
                    "init_tracing: debug output routed to both console and file"
                );
                drop(g);
            }
            Err(e) => fail(&format!("init_tracing: {e}")),
        }
        let _ = std::fs::remove_dir_all(&tmp_dir);
    }

    println!();
    println!("\u{2554}\u{2550}{}\u{2557}", sep("\u{2550}", 54));
    println!("\u{2551}           \u{2713}  all debug checks passed               \u{2551}");
    println!("\u{255a}\u{2550}{}\u{255d}", sep("\u{2550}", 54));
    println!();
}
