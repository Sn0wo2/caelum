#[cfg(feature = "nerd")]
use nerd_font_symbols::{fa, ple};
use owo_colors::Style;

#[derive(Clone, Copy, Debug)]
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

#[derive(Clone, Copy, Debug)]
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
            time_bracket_open: "「",
            time_bracket_close: "」",
            separator: "│",
            arrow: "❯",
            span_delimiter: "┇",
            span_join: "·",
        }
    }

    #[cfg(feature = "nerd")]
    pub const fn nerd() -> Self {
        Self {
            bracket_open: ple::PLE_LEFT_HALF_CIRCLE_THICK,
            bracket_close: ple::PLE_RIGHT_HALF_CIRCLE_THICK,
            time_bracket_open: "「",
            time_bracket_close: "」",
            separator: "\u{2502}",
            arrow: fa::FA_ARROW_RIGHT,
            span_delimiter: fa::FA_CODE_MERGE,
            span_join: fa::FA_ANGLE_RIGHT,
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
pub struct Theme {
    pub accent: Style,
    pub secondary: Style,
    pub text: Style,
    pub error: (u8, u8, u8),
    pub warn: (u8, u8, u8),
    pub info: (u8, u8, u8),
    pub debug: (u8, u8, u8),
    pub trace: (u8, u8, u8),
}

impl Theme {
    pub const fn trans_flag() -> Self {
        Self {
            accent: Style::new().truecolor(91, 206, 250),
            secondary: Style::new().truecolor(245, 169, 184),
            text: Style::new().truecolor(255, 255, 255),
            error: (255, 85, 85),
            warn: (255, 200, 60),
            info: (91, 206, 250),
            debug: (245, 169, 184),
            trace: (240, 240, 240),
        }
    }

    pub const fn monokai() -> Self {
        Self {
            accent: Style::new().truecolor(102, 217, 239),
            secondary: Style::new().truecolor(249, 38, 114),
            text: Style::new().truecolor(248, 248, 242),
            error: (255, 85, 85),
            warn: (255, 200, 60),
            info: (102, 217, 239),
            debug: (249, 38, 114),
            trace: (180, 180, 180),
        }
    }

    pub const fn dracula() -> Self {
        Self {
            accent: Style::new().truecolor(139, 233, 253),
            secondary: Style::new().truecolor(255, 121, 198),
            text: Style::new().truecolor(248, 248, 242),
            error: (255, 85, 85),
            warn: (255, 200, 60),
            info: (139, 233, 253),
            debug: (255, 121, 198),
            trace: (180, 180, 180),
        }
    }
    pub const fn nord() -> Self {
        Self {
            accent: Style::new().truecolor(136, 192, 208),
            secondary: Style::new().truecolor(163, 190, 140),
            text: Style::new().truecolor(216, 222, 233),
            error: (191, 97, 106),
            warn: (235, 203, 139),
            info: (136, 192, 208),
            debug: (163, 190, 140),
            trace: (180, 180, 180),
        }
    }

    pub const fn catppuccin_mocha() -> Self {
        Self {
            accent: Style::new().truecolor(137, 180, 250),
            secondary: Style::new().truecolor(203, 166, 247),
            text: Style::new().truecolor(205, 214, 244),
            error: (243, 139, 168),
            warn: (249, 226, 175),
            info: (137, 180, 250),
            debug: (203, 166, 247),
            trace: (180, 180, 180),
        }
    }

    pub const fn gruvbox() -> Self {
        Self {
            accent: Style::new().truecolor(131, 165, 152),
            secondary: Style::new().truecolor(254, 128, 25),
            text: Style::new().truecolor(235, 219, 178),
            error: (251, 73, 52),
            warn: (250, 189, 47),
            info: (131, 165, 152),
            debug: (254, 128, 25),
            trace: (180, 180, 180),
        }
    }

    pub const fn one_dark() -> Self {
        Self {
            accent: Style::new().truecolor(97, 175, 239),
            secondary: Style::new().truecolor(198, 120, 221),
            text: Style::new().truecolor(171, 178, 191),
            error: (224, 108, 117),
            warn: (229, 192, 123),
            info: (97, 175, 239),
            debug: (198, 120, 221),
            trace: (180, 180, 180),
        }
    }

    pub const fn tokyo_night() -> Self {
        Self {
            accent: Style::new().truecolor(122, 162, 247),
            secondary: Style::new().truecolor(187, 154, 247),
            text: Style::new().truecolor(192, 202, 245),
            error: (247, 118, 142),
            warn: (224, 175, 104),
            info: (122, 162, 247),
            debug: (187, 154, 247),
            trace: (180, 180, 180),
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
