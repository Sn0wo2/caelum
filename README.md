# acta

[![Rust](https://img.shields.io/badge/rust-%3E%3D1.85-orange?style=flat-square&logo=rust&logoColor=white&labelColor=1a1b27)](https://www.rust-lang.org)
[![Edition](https://img.shields.io/badge/edition-2024-blue?style=flat-square&logo=rust&logoColor=white&labelColor=1a1b27)](https://doc.rust-lang.org/edition-guide/)
[![DeepWiki](https://img.shields.io/badge/DeepWiki-acta-2ea44f?style=flat-square&logo=gitbook&logoColor=white&labelColor=1a1b27)](https://deepwiki.com/Sn0wo2/acta)

[![CI](https://github.com/Sn0wo2/acta/actions/workflows/ci.yml/badge.svg)](https://github.com/Sn0wo2/acta/actions/workflows/ci.yml)
[![Release](https://github.com/Sn0wo2/acta/actions/workflows/release.yml/badge.svg)](https://github.com/Sn0wo2/acta/actions/workflows/release.yml)

> Make Tracing Great Again.

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
| -------------- | ------------------ | -------------------------------------------------------------------------------------------------------- |
| `unicode`      | Yes                | Uses the Unicode icon set unless `nerd` selects Nerd Font icons.                                         |
| `file`         | Yes                | Enables `init_tracing`, `TracingGuard`, `build_file_layer`, and file logging through `tracing-appender`. |
| `compress`     | No                 | Enables `Rotation::Compress` for gzip-compressing old log files.                                         |
| `serde`        | No                 | Adds `Serialize` / `Deserialize` support for config types.                                               |
| `nerd`         | No                 | Enables Nerd Font icons through `Icons::nerd()` and uses them by default.                                |
| `custom-async` | No                 | Enables Tokio-backed async console writers and exports `AsyncWriter` helpers.                            |
| `native-async` | No                 | Enables non-blocking console writers backed by `tracing-appender`.                                       |
| `async`        | No                 | Enables both `custom-async` and `native-async`.                                                          |

If you disable default features, `init_tracing` is unavailable unless the `file` feature is enabled.

## Configuration

`LoggingConfig::default()` uses:

- **Level**: `Level::Info`
- **Console**: enabled with `Format::Compact`
- **Console writer**: `Writer::Stdout`
- **ANSI colors**: enabled
- **Path and span display**: enabled
- **File logging**: disabled

```rust
use acta::{
    init_tracing, ConsoleConfig, Format, Level, LoggingConfig, Result, Writer,
};

fn main() -> Result<()> {
    let config = LoggingConfig {
        level: Level::Debug,
        console: Some(ConsoleConfig {
            format: Format::Compact,
            ansi: true,
            writer: Writer::Stdout,
            show_path: true,
            show_spans: true,
            time_format: Some("%Y-%m-%d %H:%M:%S".to_string()),
            ..Default::default()
        }),
        ..Default::default()
    };

    let _guard = init_tracing(&config)?;

    Ok(())
}
```

## Console formats

| Format            | Description                                                        |
| ----------------- | ------------------------------------------------------------------ |
| `Format::Compact` | Default themed formatter with optional path and span display.      |
| `Format::Pretty`  | `tracing-subscriber` pretty formatter with file and line metadata. |
| `Format::Json`    | Flattened JSON events without ANSI colors.                         |

## File logging

File logging is available with the `file` feature, which is enabled by default. File logs are written as flattened JSON
events.

```rust
use acta::{
    init_tracing, ConsoleConfig, FileConfig, Level, LoggingConfig, Result, Rotation,
};
use std::path::PathBuf;

fn main() -> Result<()> {
    let config = LoggingConfig {
        level: Level::Info,
        console: Some(ConsoleConfig::default()),
        file: Some(FileConfig {
            path: PathBuf::from("logs/app.log"),
            rotation: Rotation::Rename,
        }),
        ..Default::default()
    };

    let _guard = init_tracing(&config)?;

    Ok(())
}
```

Supported rotation modes:

| Mode                 | Description                                                                             |
| -------------------- | --------------------------------------------------------------------------------------- |
| `Rotation::None`     | Keeps the existing log file.                                                            |
| `Rotation::Rename`   | Renames the existing log file with a timestamp before opening a new one.                |
| `Rotation::Compress` | Compresses the existing log file to gzip before opening a new one. Requires `compress`. |

## Filter directives

acta uses `tracing-subscriber` `EnvFilter` directive syntax for startup filters and runtime reloads.

```rust
use acta::{Level, LoggingConfig};

let config = LoggingConfig {
    level: Level::Custom("info,my_crate=debug,my_crate::db=trace".to_owned()),
    ..Default::default()
};
```

You can change filters after initialization through `ReloadHandle`.

```rust
use acta::{init_tracing, Filter, Level, LoggingConfig, Result};

fn main() -> Result<()> {
    let guard = init_tracing(&LoggingConfig::default())?;

    guard.reload_handle_mut().set_level(Level::Debug)?;
    guard
        .reload_handle_mut()
        .set_target_level("my_crate", Level::Trace)?;
    guard.reload_handle_mut().remove_target_level("my_crate")?;
    guard.reload_handle_mut().reload("info,my_crate=trace")?;
    guard.reload_handle_mut().set_filter(
        Filter::new(Level::Warn).with_target("my_crate", Level::Debug),
    )?;

    Ok(())
}
```

`RUST_LOG` is not read automatically. If you want to use it, pass its value into `Level::Custom`.

```rust
use acta::{Level, LoggingConfig};

let directive = std::env::var("RUST_LOG").unwrap_or_else(|_| "info".to_string());
let config = LoggingConfig {
    level: Level::Custom(directive),
    ..Default::default()
};
```

## Custom formatter

`AnsiFormatter` powers `Format::Compact` and can be customized through builder methods.

```rust
use acta::{AnsiFormatter, Icons, LevelLabels, Theme};

let formatter = AnsiFormatter::new()
    .with_theme(Theme::tokyo_night())
    .with_icons(Icons::unicode())
    .with_labels(LevelLabels::long())
    .with_time_format("%H:%M:%S")
    .with_show_path(true)
    .with_show_spans(true);
```

The default path width is generated at build time.

## Themes

| Theme                       | Description      |
| --------------------------- | ---------------- |
| `Theme::trans_flag()`       | Default          |
| `Theme::monokai()`          | Monokai          |
| `Theme::dracula()`          | Dracula          |
| `Theme::nord()`             | Nord             |
| `Theme::catppuccin_mocha()` | Catppuccin Mocha |
| `Theme::gruvbox()`          | Gruvbox          |
| `Theme::one_dark()`         | One Dark         |
| `Theme::tokyo_night()`      | Tokyo Night      |

Create custom themes from RGB values:

```rust
use acta::Theme;

let custom = Theme::new(
    (91, 206, 250),   // accent
    (245, 169, 184),  // secondary
    (255, 255, 255),  // text
    (255, 85, 85),    // error
    (255, 200, 60),   // warn
    (91, 206, 250),   // info
    (245, 169, 184),  // debug
    (240, 240, 240),  // trace
);
```

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

Custom icons and labels:

```rust
use acta::{Icons, LevelLabels};

let custom_icons = Icons::custom("[", "]", "{", "}", "|", ">", "->", "·");
let custom_labels = LevelLabels::custom("ERR", "WRN", "INF", "DBG", "TRC");
```

## Runtime style reload

`ReloadHandle` can reload themes, icons, and labels via `with_style`. `init_tracing` configures runtime filter
reloads, but not runtime style reloads.

For style reloads, build the subscriber manually:

```rust
use acta::{
    build_console_layer_with, build_reload_filter, AnsiFormatter, ConsoleConfig, Level, Result,
    Theme,
};
use tracing_subscriber::prelude::*;

fn main() -> Result<()> {
    let formatter = AnsiFormatter::new().with_theme(Theme::monokai());
    let style = *formatter.style_config();
    let console_layer = build_console_layer_with(&ConsoleConfig::default(), &formatter);
    let (filter_layer, mut reload_handle) = build_reload_filter(Level::Info, style);

    let subscriber = tracing_subscriber::registry()
        .with(console_layer)
        .with(filter_layer);
    tracing::subscriber::set_global_default(subscriber)?;

    reload_handle.with_style(|s| s.theme = Theme::dracula());

    Ok(())
}
```

This low-level setup requires adding `tracing-subscriber` as a direct dependency.

## Async console writers

With `custom-async`, `native-async`, or `async`, `Writer` gains async stdout and stderr variants.

`AsyncMode::Custom` uses Tokio, so your application must run inside a Tokio runtime. If you use `#[tokio::main]`,
add Tokio as a direct dependency with the required runtime and macro features.

```rust
use acta::{
    init_tracing, AsyncMode, ConsoleConfig, LoggingConfig, Result, Writer,
};

#[tokio::main]
async fn main() -> Result<()> {
    let config = LoggingConfig {
        console: Some(ConsoleConfig {
            writer: Writer::AsyncStdout(AsyncMode::Custom),
            ..Default::default()
        }),
        ..Default::default()
    };

    let _guard = init_tracing(&config)?;

    Ok(())
}
```

`AsyncMode::Native` uses `tracing-appender` non-blocking writers.
