use ansi_colours::ansi256_from_rgb;
use owo_colors::{DynColors, Style as OwoStyle, XtermColors};

use crate::color::mapping::rgb_to_ansi16;
use crate::config::ColorDepth;

pub fn rgb_to_owo(rgb: (u8, u8, u8), depth: ColorDepth) -> OwoStyle {
    let r = rgb.0;
    let g = rgb.1;
    let b = rgb.2;


    match depth {
        ColorDepth::TrueColor => OwoStyle::new().color(DynColors::Rgb(r, g, b)),
        ColorDepth::Ansi256 => {
            let idx = ansi256_from_rgb((r, g, b));
            OwoStyle::new().color(DynColors::Xterm(XtermColors::from(idx)))
        }
        ColorDepth::Ansi16 => {
            let c = rgb_to_ansi16(r, g, b);
            OwoStyle::new().color(DynColors::Ansi(c))
        }
        ColorDepth::NoColor => OwoStyle::new(),
    }
}

/// Dims the foreground by dividing brightness by 4 to ensure high text legibility against the background.
pub fn rgb_to_owo_on(r: u8, g: u8, b: u8, depth: ColorDepth) -> OwoStyle {
    let dr = r >> 2;
    let dg = g >> 2;
    let db = b >> 2;
    match depth {
        ColorDepth::TrueColor => OwoStyle::new()
            .color(DynColors::Rgb(dr, dg, db))
            .on_color(DynColors::Rgb(r, g, b)),
        ColorDepth::Ansi256 => {
            OwoStyle::new()
                .color(DynColors::Xterm(XtermColors::from(ansi256_from_rgb((dr, dg, db)))))
                .on_color(DynColors::Xterm(XtermColors::from(ansi256_from_rgb((r, g, b)))))
        }
        ColorDepth::Ansi16 => {
            OwoStyle::new()
                .color(DynColors::Ansi(rgb_to_ansi16(dr, dg, db)))
                .on_color(DynColors::Ansi(rgb_to_ansi16(r, g, b)))
        }
        ColorDepth::NoColor => OwoStyle::new(),
    }
}


pub fn theme_fg_dimmed(rgb: (u8, u8, u8), depth: ColorDepth) -> OwoStyle {
    rgb_to_owo((rgb.0 >> 2, rgb.1 >> 2, rgb.2 >> rgb.2 >> 2), depth)
}
