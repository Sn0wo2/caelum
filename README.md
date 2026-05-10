# acta

[![Rust](https://img.shields.io/badge/rust-1.85+-orange.svg)](https://www.rust-lang.org)
[![Edition](https://img.shields.io/badge/edition-2024-blue.svg)](https://doc.rust-lang.org/edition-guide/)

A customizable logging library for Rust

## Installation

```bash
cargo add acta
```

The default feature set enables Unicode console output and file logging.

```bash
cargo add acta --features serde,compress,nerd,async
```

## Quick start

```rust
use acta::{init_tracing, LoggingConfig, Result};

fn main() -> Result<()> {
    let _guard = init_tracing(&LoggingConfig::default())?;

    tracing::info!("Hello, acta!");
    tracing::debug!(user = "alice", "User logged in");

    Ok(())
}
```

Keep the returned guard alive for as long as logging is needed. Dropping it stops file logging.

## Features

| Feature        | Enabled by default | Description                                                                                              |
|----------------|--------------------|----------------------------------------------------------------------------------------------------------|
| `unicode`      | Yes                | Uses the Unicode icon set unless `nerd` selects Nerd Font icons.                                         |
| `file`         | Yes                | Enables `init_tracing`, `TracingGuard`, `build_file_layer`, and file logging through `tracing-appender`. |
| `compress`     | No                 | Enables `LogRotation::Compress` for gzip-compressing old log files.                                      |
| `serde`        | No                 | Adds `Serialize` / `Deserialize` support for config types.                                               |
| `nerd`         | No                 | Enables Nerd Font icons through `Icons::nerd()` and uses them by default.                                |
| `custom-async` | No                 | Enables Tokio-backed async console writers and exports `AsyncWriter` helpers.                            |
| `native-async` | No                 | Enables non-blocking console writers backed by `tracing-appender`.                                       |
| `async`        | No                 | Enables both `custom-async` and `native-async`.                                                          |

If you disable default features, `init_tracing` is unavailable unless the `file` feature is enabled.

## Configuration

`LoggingConfig::default()` uses:

- **Level**: `LogLevel::Info`
- **Console**: enabled with `LogFormat::Compact`
- **Console writer**: `ConsoleWriter::Stdout`
- **ANSI colors**: enabled
- **Path and span display**: enabled
- **File logging**: disabled

```rust
use acta::{
    init_tracing, ConsoleConfig, ConsoleWriter, LogFormat, LogLevel, LoggingConfig, Result,
};

fn main() -> Result<()> {
    let config = LoggingConfig {
        level: LogLevel::Debug,
        console: Some(ConsoleConfig {
            format: LogFormat::Compact,
            ansi: true,
            writer: ConsoleWriter::Stdout,
            show_path: true,
            show_spans: true,
            time_format: Some("%Y-%m-%d %H:%M:%S".to_string()),
        }),
        file: None,
    };

    let _guard = init_tracing(&config)?;

    Ok(())
}
```

## Console formats

| Format               | Description                                                        |
|----------------------|--------------------------------------------------------------------|
| `LogFormat::Compact` | Default themed formatter with optional path and span display.      |
| `LogFormat::Pretty`  | `tracing-subscriber` pretty formatter with file and line metadata. |
| `LogFormat::Json`    | Flattened JSON events without ANSI colors.                         |

## File logging

File logging is available with the `file` feature, which is enabled by default. File logs are written as flattened JSON
events.

```rust
use acta::{
    init_tracing, ConsoleConfig, FileLoggingConfig, LogLevel, LogRotation, LoggingConfig, Result,
};
use std::path::PathBuf;

fn main() -> Result<()> {
    let config = LoggingConfig {
        level: LogLevel::Info,
        console: Some(ConsoleConfig::default()),
        file: Some(FileLoggingConfig {
            path: PathBuf::from("logs/app.log"),
            rotation: LogRotation::Rename,
        }),
    };

    let _guard = init_tracing(&config)?;

    Ok(())
}
```

Supported rotation modes:

| Mode                    | Description                                                                             |
|-------------------------|-----------------------------------------------------------------------------------------|
| `LogRotation::None`     | Keeps the existing log file.                                                            |
| `LogRotation::Rename`   | Renames the existing log file with a timestamp before opening a new one.                |
| `LogRotation::Compress` | Compresses the existing log file to gzip before opening a new one. Requires `compress`. |

acta uses `tracing-subscriber` `EnvFilter` directive syntax for startup filters and runtime reloads.

```rust
use acta::{FilterDirective, LogLevel, LoggingConfig};

let config = LoggingConfig {
level: LogLevel::Custom(FilterDirective::new(
"info,my_crate=debug,my_crate::db=trace",
)),
..Default::default ()
};
```

You can change filters after initialization through `ReloadHandle`.

```rust
use acta::{init_tracing, LogFilter, LogLevel, LoggingConfig, Result};

fn main() -> Result<()> {
    let guard = init_tracing(&LoggingConfig::default())?;

    guard.reload_handle.set_level(LogLevel::Debug)?;
    guard
        .reload_handle
        .set_target_level("my_crate", LogLevel::Trace)?;
    guard.reload_handle.remove_target_level("my_crate")?;
    guard.reload_handle.reload("info,my_crate=trace")?;
    guard.reload_handle.set_filter(
        LogFilter::new(LogLevel::Warn).with_target_level("my_crate", LogLevel::Debug),
    )?;

    Ok(())
}
```

`RUST_LOG` is not read automatically. If you want to use it, pass its value into `LogLevel::Custom`.

```rust
use acta::{FilterDirective, LogLevel, LoggingConfig};

let directive = std::env::var("RUST_LOG").unwrap_or_else( | _ | "info".to_string());
let config = LoggingConfig {
level: LogLevel::Custom(FilterDirective::new(directive)),
..Default::default ()
};
```

## Custom formatter

`AnsiFormatter` powers `LogFormat::Compact` and can be customized through builder methods.

```rust
use acta::{AnsiFormatter, Icons, LevelLabels, Theme};

let formatter = AnsiFormatter::new()
.with_theme(Theme::tokyo_night())
.with_icons(Icons::unicode())
.with_labels(LevelLabels::long())
.with_time_format("%H:%M:%S")
.with_path_width(40)
.with_show_path(true)
.with_show_spans(true);
```

The default path width is generated at build time. Override it per formatter with `AnsiFormatter::with_path_width`.

## Themes

| Theme                       | Description      |
|-----------------------------|------------------|
| `Theme::trans_flag()`       | Default          |
| `Theme::monokai()`          | Monokai          |
| `Theme::dracula()`          | Dracula          |
| `Theme::nord()`             | Nord             |
| `Theme::catppuccin_mocha()` | Catppuccin Mocha |
| `Theme::gruvbox()`          | Gruvbox          |
| `Theme::one_dark()`         | One Dark         |
| `Theme::tokyo_night()`      | Tokyo Night      |

## Icons and labels

```rust
use acta::{Icons, LevelLabels};

let unicode_icons = Icons::unicode();
let short_labels = LevelLabels::short();
let long_labels = LevelLabels::long();
```

With the `nerd` feature enabled:

```rust
use acta::Icons;

let nerd_icons = Icons::nerd();
```

## Runtime style reload

`ReloadHandle` can reload themes, icons, and labels only when it is created with a shared `StyleConfig`. `init_tracing`
configures runtime filter reloads, but not runtime style reloads.

For style reloads, build the subscriber manually:

```rust
use acta::{
    build_console_layer_with, build_reload_filter, AnsiFormatter, ConsoleConfig, LogLevel, Result,
    Theme,
};
use tracing_subscriber::prelude::*;

fn main() -> Result<()> {
    let formatter = AnsiFormatter::new().with_theme(Theme::monokai());
    let style = formatter.style_config();
    let console_layer = build_console_layer_with(&ConsoleConfig::default(), formatter);
    let (filter_layer, reload_handle) = build_reload_filter(&LogLevel::Info, Some(style));

    let subscriber = tracing_subscriber::registry()
        .with(console_layer)
        .with(filter_layer);
    tracing::subscriber::set_global_default(subscriber)?;

    reload_handle.set_theme(Theme::dracula())?;

    Ok(())
}
```

This low-level setup requires adding `tracing-subscriber` as a direct dependency.

## Async console writers

With `custom-async`, `native-async`, or `async`, `ConsoleWriter` gains async stdout and stderr variants.

`AsyncWriterMode::Custom` uses Tokio, so your application must run inside a Tokio runtime. If you use `#[tokio::main]`,
add Tokio as a direct dependency with the required runtime and macro features.

```rust
use acta::{
    init_tracing, AsyncWriterMode, ConsoleConfig, ConsoleWriter, LoggingConfig, Result,
};

#[tokio::main]
async fn main() -> Result<()> {
    let config = LoggingConfig {
        console: Some(ConsoleConfig {
            writer: ConsoleWriter::AsyncStdout(AsyncWriterMode::Custom),
            ..Default::default()
        }),
        ..Default::default()
    };

    let _guard = init_tracing(&config)?;

    Ok(())
}
```

`AsyncWriterMode::Native` uses `tracing-appender` non-blocking writers.

