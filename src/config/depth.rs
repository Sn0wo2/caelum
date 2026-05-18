use supports_color::Stream;

use super::{ColorDepth, Writer};

/// Detect terminal color depth for the given output stream.
///
/// Priority (from `supports-color` crate, which handles NO_COLOR/FORCE_COLOR/CLICOLOR_FORCE/TTY):
/// 1. If NO_COLOR is set → NoColor (handled by supports-color)
/// 2. If FORCE_COLOR is set → map to color depth (handled by supports-color)
/// 3. If stream is not a TTY → NoColor (handled by supports-color)
/// 4. Check `supports_color::ColorLevel` for 16m/256/basic support
/// 5. Default → NoColor
pub fn detect(writer: &Writer) -> ColorDepth {
    let stream = match writer {
        Writer::Stdout => Stream::Stdout,
        Writer::Stderr => Stream::Stderr,
        #[cfg(any(feature = "custom-async", feature = "native-async"))]
        Writer::AsyncStdout(_) => Stream::Stdout,
        #[cfg(any(feature = "custom-async", feature = "native-async"))]
        Writer::AsyncStderr(_) => Stream::Stderr,
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

/// Detect whether the terminal supports Nerd Font symbols.
///
/// Checks the `NERD_FONT` environment variable (community standard
/// used by Starship, lazygit, etc). Any non-empty value other than
/// `"0"` or `"false"` is treated as true.
pub fn detect_nerd() -> bool {
    std::env::var("NERD_FONT").is_ok_and(|v| !v.is_empty() && v != "0" && v != "false")
}
