use owo_colors::{DynColors, Style as OwoStyle, XtermColors};

use crate::color::mapping::{rgb_to_ansi16, rgb_to_ansi256};
use crate::config::ColorDepth;

/// Create an OwoStyle with foreground color set to the given RGB,
/// adapted to the terminal color depth.
pub fn rgb_to_owo(r: u8, g: u8, b: u8, depth: ColorDepth) -> OwoStyle {
    match depth {
        ColorDepth::TrueColor => OwoStyle::new().color(DynColors::Rgb(r, g, b)),
        ColorDepth::Ansi256 => {
            let idx = rgb_to_ansi256(r, g, b);
            OwoStyle::new().color(DynColors::Xterm(XtermColors::from(idx)))
        }
        ColorDepth::Ansi16 => {
            let c = rgb_to_ansi16(r, g, b);
            OwoStyle::new().color(DynColors::Ansi(c))
        }
        ColorDepth::NoColor => OwoStyle::new(),
    }
}

/// Create an OwoStyle with foreground AND background (on_color) set, depth-aware.
/// The foreground uses a dimmed variant (right-shift R,G,B by 2).
pub fn rgb_to_owo_on(r: u8, g: u8, b: u8, depth: ColorDepth) -> OwoStyle {
    let dr = r >> 2;
    let dg = g >> 2;
    let db = b >> 2;
    match depth {
        ColorDepth::TrueColor => OwoStyle::new()
            .color(DynColors::Rgb(dr, dg, db))
            .on_color(DynColors::Rgb(r, g, b)),
        ColorDepth::Ansi256 => {
            let fg_idx = rgb_to_ansi256(dr, dg, db);
            let bg_idx = rgb_to_ansi256(r, g, b);
            OwoStyle::new()
                .color(DynColors::Xterm(XtermColors::from(fg_idx)))
                .on_color(DynColors::Xterm(XtermColors::from(bg_idx)))
        }
        ColorDepth::Ansi16 => {
            let fg = rgb_to_ansi16(dr, dg, db);
            let bg = rgb_to_ansi16(r, g, b);
            OwoStyle::new()
                .color(DynColors::Ansi(fg))
                .on_color(DynColors::Ansi(bg))
        }
        ColorDepth::NoColor => OwoStyle::new(),
    }
}

/// Create an OwoStyle for a Theme color (accent/secondary/text) at the given color depth.
/// This replaces the old pattern of `theme.accent.style(...)`.
pub fn theme_fg(rgb: (u8, u8, u8), depth: ColorDepth) -> OwoStyle {
    rgb_to_owo(rgb.0, rgb.1, rgb.2, depth)
}

/// Create a dimmed OwoStyle for a Theme color at the given color depth.
/// This replaces the old pattern of `theme.accent.dimmed().style(...)`.
pub fn theme_fg_dimmed(rgb: (u8, u8, u8), depth: ColorDepth) -> OwoStyle {
    let dr = rgb.0 >> 2;
    let dg = rgb.1 >> 2;
    let db = rgb.2 >> 2;
    rgb_to_owo(dr, dg, db, depth)
}
