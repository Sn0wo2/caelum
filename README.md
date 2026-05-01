# sage-trace

[![Rust](https://img.shields.io/badge/rust-1.85+-orange.svg)](https://www.rust-lang.org)
[![Edition](https://img.shields.io/badge/edition-2024-blue.svg)](https://doc.rust-lang.org/edition-guide/)

一个基于 `tracing` 的 Rust 日志库，支持彩色控制台输出、文件日志和运行时日志级别调整。

## 安装

```toml
[dependencies]
sage-trace = "0.1"
tracing = "0.1"
anyhow = "1"
```

## 快速开始

```rust
use sage_trace::{init_tracing, LoggingConfig};

fn main() -> anyhow::Result<()> {
    // 使用默认配置初始化
    let _guard = init_tracing(&LoggingConfig::default())?;

    tracing::info!("Hello, sage-trace!");
    tracing::debug!(user = "alice", "User logged in");

    Ok(())
}
```

## 常用用法

### 基础配置

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

### 文件日志

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
        rotation: LogRotation::Compress,  // 支持 None, Rename, Compress
    }),
};

let guard = init_tracing(&config)?;
// guard 必须在程序运行期间保持存活
```

### 自定义主题

```rust
use sage_trace::{AnsiFormatter, Theme, Icons};

let formatter = AnsiFormatter::new()
    .with_theme(Theme::monokai())
    .with_icons(Icons::unicode())
    .with_show_path(true)
    .with_show_spans(true);
```

### 运行时修改日志级别

```rust
use sage_trace::{init_tracing, LogLevel, LoggingConfig};

let guard = init_tracing(&LoggingConfig::default())?;

guard.reload_handle.set_level(LogLevel::Debug)?;
guard
    .reload_handle
    .set_target_level("my_crate", LogLevel::Trace)?;
```

## 主题预览

| 主题 | 描述                  |
|------|---------------------|
| `Theme::trans_flag()` | 蓝粉白（默认）             |
| `Theme::monokai()` | Monokai 编辑器主题       |
| `Theme::dracula()` | Dracula 暗色主题        |
| `Theme::nord()` | Nord 极地主题           |
| `Theme::catppuccin_mocha()` | Catppuccin Mocha 主题 |
| `Theme::gruvbox()` | Gruvbox 复古主题        |
| `Theme::one_dark()` | One Dark 主题         |
| `Theme::tokyo_night()` | Tokyo Night 主题      |

## 环境变量过滤

`sage-trace` 支持通过 `RUST_LOG` 环境变量进行细粒度控制：

```bash
RUST_LOG=info,my_crate=debug,my_crate::module=trace cargo run
```

## 编译时配置

设置路径显示的最大宽度（默认 28）：

```bash
SAGE_TRACE_MAX_PATH_WIDTH=40 cargo build
```
