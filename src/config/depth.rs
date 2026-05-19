use supports_color::Stream;

use super::{ColorDepth, WriterTarget};

pub fn detect(target: &WriterTarget) -> ColorDepth {
    let stream = match target {
        WriterTarget::Stdout => Stream::Stdout,
        WriterTarget::Stderr => Stream::Stderr,
        #[cfg(feature = "file")]
        WriterTarget::File { .. } => return ColorDepth::NoColor,
        #[cfg(any(feature = "custom-async", feature = "native-async"))]
        WriterTarget::AsyncStdout(_) => Stream::Stdout,
        #[cfg(any(feature = "custom-async", feature = "native-async"))]
        WriterTarget::AsyncStderr(_) => Stream::Stderr,
    };

    if let Some(level) = supports_color::on_cached(stream) {
        if level.has_16m {
            return ColorDepth::TrueColor;
        }
        if level.has_256 {
            return ColorDepth::Ansi256;
        }
        if level.has_basic {
            return ColorDepth::Ansi16;
        }
    }

    ColorDepth::NoColor
}

pub fn detect_nerd() -> bool {
    std::env::var("NERD_FONT").is_ok_and(|v| !v.is_empty() && v != "0" && v != "false")
}
