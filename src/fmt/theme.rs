#[cfg(feature = "nerd")]
use nerd_font_symbols::{cod, fa, ple};
use owo_colors::Style;

#[derive(Clone, Copy, Debug, PartialEq)]
pub struct LevelLabels {
    pub error: &'static str,
    pub warn: &'static str,
    pub info: &'static str,
    pub debug: &'static str,
    pub trace: &'static str,
}

impl LevelLabels {
    pub const fn short() -> Self {
        Self {
            error: "E",
            warn: "W",
            info: "I",
            debug: "D",
            trace: "T",
        }
    }

    pub const fn long() -> Self {
        Self {
            error: "ERROR",
            warn: " WARN",
            info: " INFO",
            debug: "DEBUG",
            trace: "TRACE",
        }
    }
}

impl Default for LevelLabels {
    fn default() -> Self {
        Self::short()
    }
}

#[derive(Clone, Copy, Debug, PartialEq)]
pub struct Icons {
    pub bracket_open: &'static str,
    pub bracket_close: &'static str,
    pub time_bracket_open: &'static str,
    pub time_bracket_close: &'static str,
    pub separator: &'static str,
    pub arrow: &'static str,
    pub span_delimiter: &'static str,
    pub span_join: &'static str,
}

impl Icons {
    pub const fn unicode() -> Self {
        Self {
            bracket_open: "[",
            bracket_close: "]",
            time_bracket_open: "\u{300c}",
            time_bracket_close: "\u{300d}",
            separator: "\u{2507}",
            arrow: "\u{276f}",
            span_delimiter: "->",
            span_join: "\u{b7}",
        }
    }

    #[cfg(feature = "nerd")]
    pub const fn nerd() -> Self {
        Self {
            bracket_open: ple::PLE_LEFT_HALF_CIRCLE_THICK,
            bracket_close: ple::PLE_RIGHT_HALF_CIRCLE_THICK,
            time_bracket_open: "\u{300c}",
            time_bracket_close: "\u{300d}",
            separator: "\u{2507}",
            arrow: fa::FA_ARROW_RIGHT,
            span_delimiter: cod::COD_EXPORT,
            span_join: fa::FA_ANGLES_RIGHT,
        }
    }

    pub fn is_nerd(&self) -> bool {
        self.bracket_open != "["
    }
}

impl Default for Icons {
    fn default() -> Self {
        #[cfg(feature = "nerd")]
        {
            Self::nerd()
        }
        #[cfg(not(feature = "nerd"))]
        {
            Self::unicode()
        }
    }
}

#[derive(Clone, Copy, Debug)]
pub struct LevelColors {
    pub rgb: (u8, u8, u8),
    pub dark: (u8, u8, u8),
    pub bg: Style,
}

impl LevelColors {
    pub fn new(rgb: (u8, u8, u8)) -> Self {
        let dark = (rgb.0 >> 1, rgb.1 >> 1, rgb.2 >> 1);
        let bg = Style::new()
            .on_truecolor(rgb.0, rgb.1, rgb.2)
            .truecolor(dark.0, dark.1, dark.2)
            .bold();
        Self { rgb, dark, bg }
    }
}

#[derive(Clone, Copy, Debug)]
pub struct Theme {
    pub accent: Style,
    pub secondary: Style,
    pub text: Style,
    pub accent_dimmed: Style,
    pub text_dimmed: Style,
    pub error: LevelColors,
    pub warn: LevelColors,
    pub info: LevelColors,
    pub debug: LevelColors,
    pub trace: LevelColors,
}

impl Theme {
    pub fn trans_flag() -> Self {
        let accent = Style::new().truecolor(91, 206, 250);
        let secondary = Style::new().truecolor(245, 169, 184);
        let text = Style::new().truecolor(255, 255, 255);
        Self {
            accent_dimmed: accent.dimmed(),
            text_dimmed: text.dimmed(),
            accent,
            secondary,
            text,
            error: LevelColors::new((255, 85, 85)),
            warn: LevelColors::new((255, 200, 60)),
            info: LevelColors::new((91, 206, 250)),
            debug: LevelColors::new((245, 169, 184)),
            trace: LevelColors::new((240, 240, 240)),
        }
    }

    pub fn monokai() -> Self {
        let accent = Style::new().truecolor(102, 217, 239);
        let secondary = Style::new().truecolor(249, 38, 114);
        let text = Style::new().truecolor(248, 248, 242);
        Self {
            accent_dimmed: accent.dimmed(),
            text_dimmed: text.dimmed(),
            accent,
            secondary,
            text,
            error: LevelColors::new((255, 85, 85)),
            warn: LevelColors::new((255, 200, 60)),
            info: LevelColors::new((102, 217, 239)),
            debug: LevelColors::new((249, 38, 114)),
            trace: LevelColors::new((180, 180, 180)),
        }
    }

    pub fn dracula() -> Self {
        let accent = Style::new().truecolor(139, 233, 253);
        let secondary = Style::new().truecolor(255, 121, 198);
        let text = Style::new().truecolor(248, 248, 242);
        Self {
            accent_dimmed: accent.dimmed(),
            text_dimmed: text.dimmed(),
            accent,
            secondary,
            text,
            error: LevelColors::new((255, 85, 85)),
            warn: LevelColors::new((255, 200, 60)),
            info: LevelColors::new((139, 233, 253)),
            debug: LevelColors::new((255, 121, 198)),
            trace: LevelColors::new((180, 180, 180)),
        }
    }

    pub fn nord() -> Self {
        let accent = Style::new().truecolor(136, 192, 208);
        let secondary = Style::new().truecolor(163, 190, 140);
        let text = Style::new().truecolor(216, 222, 233);
        Self {
            accent_dimmed: accent.dimmed(),
            text_dimmed: text.dimmed(),
            accent,
            secondary,
            text,
            error: LevelColors::new((191, 97, 106)),
            warn: LevelColors::new((235, 203, 139)),
            info: LevelColors::new((136, 192, 208)),
            debug: LevelColors::new((163, 190, 140)),
            trace: LevelColors::new((180, 180, 180)),
        }
    }

    pub fn catppuccin_mocha() -> Self {
        let accent = Style::new().truecolor(137, 180, 250);
        let secondary = Style::new().truecolor(203, 166, 247);
        let text = Style::new().truecolor(205, 214, 244);
        Self {
            accent_dimmed: accent.dimmed(),
            text_dimmed: text.dimmed(),
            accent,
            secondary,
            text,
            error: LevelColors::new((243, 139, 168)),
            warn: LevelColors::new((249, 226, 175)),
            info: LevelColors::new((137, 180, 250)),
            debug: LevelColors::new((203, 166, 247)),
            trace: LevelColors::new((180, 180, 180)),
        }
    }

    pub fn gruvbox() -> Self {
        let accent = Style::new().truecolor(131, 165, 152);
        let secondary = Style::new().truecolor(254, 128, 25);
        let text = Style::new().truecolor(235, 219, 178);
        Self {
            accent_dimmed: accent.dimmed(),
            text_dimmed: text.dimmed(),
            accent,
            secondary,
            text,
            error: LevelColors::new((251, 73, 52)),
            warn: LevelColors::new((250, 189, 47)),
            info: LevelColors::new((131, 165, 152)),
            debug: LevelColors::new((254, 128, 25)),
            trace: LevelColors::new((180, 180, 180)),
        }
    }

    pub fn one_dark() -> Self {
        let accent = Style::new().truecolor(97, 175, 239);
        let secondary = Style::new().truecolor(198, 120, 221);
        let text = Style::new().truecolor(171, 178, 191);
        Self {
            accent_dimmed: accent.dimmed(),
            text_dimmed: text.dimmed(),
            accent,
            secondary,
            text,
            error: LevelColors::new((224, 108, 117)),
            warn: LevelColors::new((229, 192, 123)),
            info: LevelColors::new((97, 175, 239)),
            debug: LevelColors::new((198, 120, 221)),
            trace: LevelColors::new((180, 180, 180)),
        }
    }

    pub fn tokyo_night() -> Self {
        let accent = Style::new().truecolor(122, 162, 247);
        let secondary = Style::new().truecolor(187, 154, 247);
        let text = Style::new().truecolor(192, 202, 245);
        Self {
            accent_dimmed: accent.dimmed(),
            text_dimmed: text.dimmed(),
            accent,
            secondary,
            text,
            error: LevelColors::new((247, 118, 142)),
            warn: LevelColors::new((224, 175, 104)),
            info: LevelColors::new((122, 162, 247)),
            debug: LevelColors::new((187, 154, 247)),
            trace: LevelColors::new((180, 180, 180)),
        }
    }
}

impl Default for Theme {
    fn default() -> Self {
        Self::trans_flag()
    }
}

#[derive(Clone, Copy, Debug, Default)]
pub struct StyleConfig {
    pub theme: Theme,
    pub icons: Icons,
    pub labels: LevelLabels,
}
