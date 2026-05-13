#[cfg(feature = "nerd")]
use nerd_font_symbols::{cod, fa, ple};
use owo_colors::Style;

#[derive(Clone, Copy, Debug, PartialEq)]
#[non_exhaustive]
pub struct LevelLabels {
    pub error: &'static str,
    pub warn: &'static str,
    pub info: &'static str,
    pub debug: &'static str,
    pub trace: &'static str,
}

impl LevelLabels {
    pub const fn custom(
        error: &'static str,
        warn: &'static str,
        info: &'static str,
        debug: &'static str,
        trace: &'static str,
    ) -> Self {
        Self {
            error,
            warn,
            info,
            debug,
            trace,
        }
    }

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
#[non_exhaustive]
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
    #[allow(clippy::too_many_arguments)]
    pub const fn custom(
        bracket_open: &'static str,
        bracket_close: &'static str,
        time_bracket_open: &'static str,
        time_bracket_close: &'static str,
        separator: &'static str,
        arrow: &'static str,
        span_delimiter: &'static str,
        span_join: &'static str,
    ) -> Self {
        Self {
            bracket_open,
            bracket_close,
            time_bracket_open,
            time_bracket_close,
            separator,
            arrow,
            span_delimiter,
            span_join,
        }
    }

    pub const fn unicode() -> Self {
        Self {
            bracket_open: "[",
            bracket_close: "]",
            time_bracket_open: "\u{300c}",
            time_bracket_close: "\u{300d}",
            separator: "\u{2507}",
            arrow: ">",
            span_delimiter: "->",
            span_join: "\u{00bb}",
        }
    }

    #[cfg(feature = "nerd")]
    pub const fn nerd() -> Self {
        Self {
            bracket_open: ple::PLE_LEFT_HALF_CIRCLE_THICK,
            bracket_close: ple::PLE_RIGHT_HALF_CIRCLE_THICK,
            time_bracket_open: ple::PLE_LEFT_HALF_CIRCLE_THIN,
            time_bracket_close: ple::PLE_RIGHT_HALF_CIRCLE_THIN,
            separator: "\u{2507}",
            arrow: fa::FA_CARET_RIGHT,
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

/// Raw RGB color values for constructing a [`Theme`].
/// Raw RGB color values for constructing a [`Theme`].
///
/// These are compile-time constants that get "baked" into a
/// runtime [`Theme`] via [`Theme::new`].

#[derive(Clone, Copy, Debug)]
pub struct ThemeRgb {
    pub accent: (u8, u8, u8),
    pub secondary: (u8, u8, u8),
    pub text: (u8, u8, u8),
    pub error: (u8, u8, u8),
    pub warn: (u8, u8, u8),
    pub info: (u8, u8, u8),
    pub debug: (u8, u8, u8),
    pub trace: (u8, u8, u8),
}

impl ThemeRgb {
    pub const fn trans_flag() -> Self {
        Self {
            accent: (91, 206, 250),
            secondary: (245, 169, 184),
            text: (255, 255, 255),
            error: (255, 85, 85),
            warn: (255, 200, 60),
            info: (91, 206, 250),
            debug: (245, 169, 184),
            trace: (240, 240, 240),
        }
    }

    pub const fn monokai() -> Self {
        Self {
            accent: (102, 217, 239),
            secondary: (249, 38, 114),
            text: (248, 248, 242),
            error: (255, 85, 85),
            warn: (255, 200, 60),
            info: (102, 217, 239),
            debug: (249, 38, 114),
            trace: (180, 180, 180),
        }
    }

    pub const fn dracula() -> Self {
        Self {
            accent: (139, 233, 253),
            secondary: (255, 121, 198),
            text: (248, 248, 242),
            error: (255, 85, 85),
            warn: (255, 200, 60),
            info: (139, 233, 253),
            debug: (255, 121, 198),
            trace: (180, 180, 180),
        }
    }

    pub const fn nord() -> Self {
        Self {
            accent: (136, 192, 208),
            secondary: (163, 190, 140),
            text: (216, 222, 233),
            error: (191, 97, 106),
            warn: (235, 203, 139),
            info: (136, 192, 208),
            debug: (163, 190, 140),
            trace: (180, 180, 180),
        }
    }

    pub const fn catppuccin_mocha() -> Self {
        Self {
            accent: (137, 180, 250),
            secondary: (203, 166, 247),
            text: (205, 214, 244),
            error: (243, 139, 168),
            warn: (249, 226, 175),
            info: (137, 180, 250),
            debug: (203, 166, 247),
            trace: (180, 180, 180),
        }
    }

    pub const fn gruvbox() -> Self {
        Self {
            accent: (131, 165, 152),
            secondary: (254, 128, 25),
            text: (235, 219, 178),
            error: (251, 73, 52),
            warn: (250, 189, 47),
            info: (131, 165, 152),
            debug: (254, 128, 25),
            trace: (180, 180, 180),
        }
    }

    pub const fn one_dark() -> Self {
        Self {
            accent: (97, 175, 239),
            secondary: (198, 120, 221),
            text: (171, 178, 191),
            error: (224, 108, 117),
            warn: (229, 192, 123),
            info: (97, 175, 239),
            debug: (198, 120, 221),
            trace: (180, 180, 180),
        }
    }

    pub const fn tokyo_night() -> Self {
        Self {
            accent: (122, 162, 247),
            secondary: (187, 154, 247),
            text: (192, 202, 245),
            error: (247, 118, 142),
            warn: (224, 175, 104),
            info: (122, 162, 247),
            debug: (187, 154, 247),
            trace: (180, 180, 180),
        }
    }
}

#[derive(Clone, Copy, Debug)]
#[non_exhaustive]
pub struct Theme {
    pub accent: Style,
    pub secondary: Style,
    pub text: Style,

    // RGB
    pub error: (u8, u8, u8),
    pub warn: (u8, u8, u8),
    pub info: (u8, u8, u8),
    pub debug: (u8, u8, u8),
    pub trace: (u8, u8, u8),
}

impl Theme {
    pub fn new(rgb: ThemeRgb) -> Self {
        Self {
            accent: Style::new().truecolor(rgb.accent.0, rgb.accent.1, rgb.accent.2),
            secondary: Style::new().truecolor(rgb.secondary.0, rgb.secondary.1, rgb.secondary.2),
            text: Style::new().truecolor(rgb.text.0, rgb.text.1, rgb.text.2),
            error: rgb.error,
            warn: rgb.warn,
            info: rgb.info,
            debug: rgb.debug,
            trace: rgb.trace,
        }
    }

    pub fn trans_flag() -> Self {
        Self::new(ThemeRgb::trans_flag())
    }

    pub fn monokai() -> Self {
        Self::new(ThemeRgb::monokai())
    }

    pub fn dracula() -> Self {
        Self::new(ThemeRgb::dracula())
    }

    pub fn nord() -> Self {
        Self::new(ThemeRgb::nord())
    }

    pub fn catppuccin_mocha() -> Self {
        Self::new(ThemeRgb::catppuccin_mocha())
    }

    pub fn gruvbox() -> Self {
        Self::new(ThemeRgb::gruvbox())
    }

    pub fn one_dark() -> Self {
        Self::new(ThemeRgb::one_dark())
    }

    pub fn tokyo_night() -> Self {
        Self::new(ThemeRgb::tokyo_night())
    }
}

impl Default for Theme {
    fn default() -> Self {
        Self::trans_flag()
    }
}

#[derive(Clone, Copy, Debug, Default)]
#[non_exhaustive]
pub struct StyleConfig {
    pub theme: Theme,
    pub icons: Icons,
    pub labels: LevelLabels,
}
