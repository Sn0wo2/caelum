use sage_trace::*;
use tracing_subscriber::prelude::*;

fn demo_logs(label: &str) {
    tracing::info!("[{label}] hello from compact");
    tracing::warn!(user = "alice", count = 42, "[{label}] something happened");
    tracing::error!("[{label}] this is an error");
    tracing::debug!("[{label}] debug message");
    let _span = tracing::info_span!("my_span", task = "demo").entered();
    tracing::trace!("[{label}] inside a span");
}

fn main() {
    println!("▸ Test 1a: Compact + Unicode icons (default)");
    {
        let console = ConsoleConfig::default();
        let formatter = AnsiFormatter::new().with_icons(Icons::unicode());
        let layer = build_console_layer_with(&console, formatter);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs("unicode"));
    }

    #[cfg(feature = "nerd")]
    println!("\n▸ Test 1b: Compact + Nerd Font icons");
    #[cfg(feature = "nerd")]
    {
        let console = ConsoleConfig::default();
        let formatter = AnsiFormatter::new().with_icons(Icons::nerd());
        let layer = build_console_layer_with(&console, formatter);
        let subscriber = tracing_subscriber::registry().with(layer);
        tracing::subscriber::with_default(subscriber, || demo_logs("nerd"));
    }

    println!("\n▸ Test 2: AnsiFormatter builder API");
    {
        let fmt = AnsiFormatter::new()
            .with_theme(Theme::monokai())
            .with_time_format("%Y-%m-%d %H:%M:%S")
            .with_path_width(40)
            .with_show_path(false)
            .with_show_spans(false);

        println!(
            "  AnsiFormatter {{ time_format: {:?}, path_width: {}, show_path: {}, show_spans: {} }}",
            fmt.time_format, fmt.path_width, fmt.show_path, fmt.show_spans
        );
    }

    println!("\n▸ Test 3: Theme presets");
    {
        let themes: &[(&str, Theme)] = &[
            ("trans_flag", Theme::trans_flag()),
            ("monokai", Theme::monokai()),
            ("dracula", Theme::dracula()),
            ("nord", Theme::nord()),
            ("catppuccin_mocha", Theme::catppuccin_mocha()),
            ("gruvbox", Theme::gruvbox()),
            ("one_dark", Theme::one_dark()),
            ("tokyo_night", Theme::tokyo_night()),
        ];
        for (name, theme) in themes {
            println!(
                "  {name}: accent={:?}, secondary={:?}, text={:?}",
                theme.accent, theme.secondary, theme.text
            );
        }
    }

    println!("\n▸ Test 3b: All themes with actual log output");
    {
        let themes: &[(&str, Theme)] = &[
            ("trans_flag", Theme::trans_flag()),
            ("monokai", Theme::monokai()),
            ("dracula", Theme::dracula()),
            ("nord", Theme::nord()),
            ("catppuccin_mocha", Theme::catppuccin_mocha()),
            ("gruvbox", Theme::gruvbox()),
            ("one_dark", Theme::one_dark()),
            ("tokyo_night", Theme::tokyo_night()),
        ];
        for (name, theme) in themes {
            println!("  [{name}]");
            let console = ConsoleConfig::default();
            let formatter = AnsiFormatter::new()
                .with_theme(theme.clone())
                .with_show_path(false);
            let layer = build_console_layer_with(&console, formatter);
            let subscriber = tracing_subscriber::registry().with(layer);
            tracing::subscriber::with_default(subscriber, || {
                tracing::info!(user = "alice", count = 42, "hello from {name}");
            });
        }
    }

    println!("\n▸ Test 4: ConsoleConfig variations");
    {
        let configs = vec![
            ("compact+stdout", ConsoleConfig::default()),
            (
                "json+stderr",
                ConsoleConfig {
                    format: LogFormat::Json,
                    writer: ConsoleWriter::Stderr,
                    ansi: false,
                    ..Default::default()
                },
            ),
            (
                "pretty+no-path",
                ConsoleConfig {
                    format: LogFormat::Pretty,
                    show_path: false,
                    show_spans: false,
                    ..Default::default()
                },
            ),
        ];
        for (label, cfg) in &configs {
            println!("  {label}: {:?}", cfg);
        }
    }

    println!("\n▸ Test 5: File logging with rotation");
    {
        let tmp_dir = std::env::temp_dir().join("sage-trace-debug");
        let _ = std::fs::create_dir_all(&tmp_dir);
        let log_path = tmp_dir.join("test.log");

        let file_config = FileLoggingConfig {
            path: log_path.clone(),
            rotation: LogRotation::Rename,
        };

        println!("  file_config: {:?}", file_config);
        println!("  log would go to: {}", log_path.display());

        std::fs::write(&log_path, b"old log content\n").ok();
        match rotate_log_file(&log_path, LogRotation::Rename) {
            Ok(()) => println!("  rotate_log_file(Rename): OK"),
            Err(e) => println!("  rotate_log_file(Rename): ERR {e}"),
        }

        #[cfg(feature = "compress")]
        {
            std::fs::write(&log_path, b"compress me\n").ok();
            match rotate_log_file(&log_path, LogRotation::Compress) {
                Ok(()) => println!("  rotate_log_file(Compress): OK"),
                Err(e) => println!("  rotate_log_file(Compress): ERR {e}"),
            }
        }

        if let Ok(entries) = std::fs::read_dir(&tmp_dir) {
            println!("  rotated files:");
            for entry in entries.flatten() {
                println!("    {}", entry.file_name().to_string_lossy());
            }
        }

        let _ = std::fs::remove_dir_all(&tmp_dir);
    }

    println!("\n▸ Test 6: LogLevel directive mapping");
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
            println!("  {:?} -> \"{}\"", level, level.as_filter_directive());
        }
    }

    println!("\n▸ Test 7: build_console_layer smoke test");
    {
        let layer = build_console_layer(&ConsoleConfig::default());
        let _ = layer;
        println!("  build_console_layer(default): OK");

        let layer = build_console_layer(&ConsoleConfig {
            format: LogFormat::Json,
            writer: ConsoleWriter::Stderr,
            ansi: false,
            ..Default::default()
        });
        let _ = layer;
        println!("  build_console_layer(json+stderr): OK");
    }

    println!("\n▸ Test 8: build_file_layer smoke test");
    {
        let tmp_dir = std::env::temp_dir().join("sage-trace-debug-file");
        let _ = std::fs::create_dir_all(&tmp_dir);

        let result = build_file_layer(&FileLoggingConfig {
            path: tmp_dir.join("app.log"),
            rotation: LogRotation::None,
        });

        match result {
            Ok(r) => {
                println!("  build_file_layer: OK, path={}", r.path.display());
                drop(r.guard);
            }
            Err(e) => println!("  build_file_layer: ERR {e}"),
        }

        let _ = std::fs::remove_dir_all(&tmp_dir);
    }

    println!("\n=== all debug checks passed ===");
}
