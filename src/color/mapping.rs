use owo_colors::AnsiColors;

use anstyle::RgbColor;
use anstyle_lossy::{palette::Palette, rgb_to_ansi};

/// Delegates to the `anstyle_lossy` crate which uses a weighted
/// Euclidean distance metric against a platform-aware default palette
/// (VGA on Unix, Windows 10 Console on Windows).
pub fn rgb_to_ansi16(r: u8, g: u8, b: u8) -> AnsiColors {
    let ansi = rgb_to_ansi(RgbColor(r, g, b), Palette::default());
    const PALETTE: [AnsiColors; 16] = [
        AnsiColors::Black,
        AnsiColors::Red,
        AnsiColors::Green,
        AnsiColors::Yellow,
        AnsiColors::Blue,
        AnsiColors::Magenta,
        AnsiColors::Cyan,
        AnsiColors::White,
        AnsiColors::BrightBlack,
        AnsiColors::BrightRed,
        AnsiColors::BrightGreen,
        AnsiColors::BrightYellow,
        AnsiColors::BrightBlue,
        AnsiColors::BrightMagenta,
        AnsiColors::BrightCyan,
        AnsiColors::BrightWhite,
    ];
    *PALETTE.get(ansi as usize).unwrap_or(&AnsiColors::White)
}
