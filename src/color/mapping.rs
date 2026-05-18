use owo_colors::AnsiColors;

use anstyle::RgbColor;
use anstyle_lossy::{palette::Palette, rgb_to_ansi};

/// Convert 24-bit RGB to the nearest ANSI 256-color index (0-255).
///
/// Delegates to the `ansi_colours` crate.
pub fn rgb_to_ansi256(r: u8, g: u8, b: u8) -> u8 {
    ansi_colours::ansi256_from_rgb((r, g, b))
}

/// Convert 24-bit RGB to the nearest 16-color ANSI variant.
///
/// Delegates to the `anstyle_lossy` crate which uses a weighted
/// Euclidean distance metric against a platform-aware default palette
/// (VGA on Unix, Windows 10 Console on Windows).
pub fn rgb_to_ansi16(r: u8, g: u8, b: u8) -> AnsiColors {
    let ansi = rgb_to_ansi(RgbColor(r, g, b), Palette::default());

    match ansi {
        anstyle::AnsiColor::Black => AnsiColors::Black,
        anstyle::AnsiColor::Red => AnsiColors::Red,
        anstyle::AnsiColor::Green => AnsiColors::Green,
        anstyle::AnsiColor::Yellow => AnsiColors::Yellow,
        anstyle::AnsiColor::Blue => AnsiColors::Blue,
        anstyle::AnsiColor::Magenta => AnsiColors::Magenta,
        anstyle::AnsiColor::Cyan => AnsiColors::Cyan,
        anstyle::AnsiColor::White => AnsiColors::White,
        anstyle::AnsiColor::BrightBlack => AnsiColors::BrightBlack,
        anstyle::AnsiColor::BrightRed => AnsiColors::BrightRed,
        anstyle::AnsiColor::BrightGreen => AnsiColors::BrightGreen,
        anstyle::AnsiColor::BrightYellow => AnsiColors::BrightYellow,
        anstyle::AnsiColor::BrightBlue => AnsiColors::BrightBlue,
        anstyle::AnsiColor::BrightMagenta => AnsiColors::BrightMagenta,
        anstyle::AnsiColor::BrightCyan => AnsiColors::BrightCyan,
        anstyle::AnsiColor::BrightWhite => AnsiColors::BrightWhite,
    }
}
