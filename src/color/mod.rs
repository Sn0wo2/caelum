//! Color depth handling, RGB→ANSI mapping, and theme color style helpers.
//!
//! Sub-modules:
//! - [`mapping`]: rgb_to_ansi256 / rgb_to_ansi16 (thin wrappers around `ansi_colours`)
//! - [`style`]: rgb_to_owo / theme_fg 等 OwoStyle 生成函数

pub mod mapping;
pub mod style;

#[cfg(test)]
mod tests;

pub use mapping::{rgb_to_ansi16, rgb_to_ansi256};
pub use style::{rgb_to_owo, rgb_to_owo_on, theme_fg, theme_fg_dimmed};
