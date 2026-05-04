# sage-trace

[![Rust](https://img.shields.io/badge/rust-1.85+-orange.svg)](https://www.rust-lang.org)
[![Edition](https://img.shields.io/badge/edition-2024-blue.svg)](https://doc.rust-lang.org/edition-guide/)

```bash
cargo add sage-trace
```

## Quick Start

```rust
use sage_trace::{init_tracing, LoggingConfig};

fn main() -> anyhow::Result<()> {
    let _guard = init_tracing(&LoggingConfig::default())?;

    tracing::info!("Hello, sage-trace!");
    tracing::debug!(user = "alice", "User logged in");

    Ok(())
}
```

## Examples

### Base config

```rust
use sage_trace::{
    init_tracing, LoggingConfig, ConsoleConfig, LogLevel, LogFormat
};

let config = LoggingConfig {
    level: LogLevel::Debug,
    console: Some(ConsoleConfig {
        format: LogFormat::Compact,
        ansi: true,
        ..Default::default()
    }),
    file: None,
};

let guard = init_tracing(&config)?;
```

### Log to file

```rust
use sage_trace::{
    init_tracing, LoggingConfig, FileLoggingConfig, LogRotation
};
use std::path::PathBuf;

let config = LoggingConfig {
    level: LogLevel::Info,
    console: Some(ConsoleConfig::default()),
    file: Some(FileLoggingConfig {
        path: PathBuf::from("logs/app.log"),
        rotation: LogRotation::Compress,  // Supports None, Rename, Compress
    }),
};

let guard = init_tracing(&config)?;
// The guard must be kept alive while the program is running
```

### Custom theme

```rust
use sage_trace::{AnsiFormatter, Theme, Icons};

let formatter = AnsiFormatter::new()
    .with_theme(Theme::monokai())
    .with_icons(Icons::unicode())
    .with_show_path(true)
    .with_show_spans(true);
```

### Edit log level in runtime

```rust
use sage_trace::{init_tracing, LogLevel, LoggingConfig};

let guard = init_tracing(&LoggingConfig::default())?;

guard.reload_handle.set_level(LogLevel::Debug)?;
guard
    .reload_handle
    .set_target_level("my_crate", LogLevel::Trace)?;
```

## Theme

| Theme                       | Description               |
|-----------------------------|------------------|
| `Theme::trans_flag()`       | Default          |
| `Theme::monokai()`          | Monokai          |
| `Theme::dracula()`          | Dracula          |
| `Theme::nord()`             | Nord             |
| `Theme::catppuccin_mocha()` | Catppuccin Mocha |
| `Theme::gruvbox()`          | Gruvbox          |
| `Theme::one_dark()`         | One Dark         |
| `Theme::tokyo_night()`      | Tokyo Night      |

## Environment filter

Supports the RUST_LOG environment variable for log level control

```bash
RUST_LOG=info,my_crate=debug,my_crate::module=trace cargo run
```

## Set at compile time

Set maximum path width(Default: 28):

```bash
SAGE_TRACE_MAX_PATH_WIDTH=40 cargo build
```
