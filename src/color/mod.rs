pub mod mapping;
pub mod style;

#[cfg(test)]
mod tests;

pub use mapping::rgb_to_ansi16;
pub use style::{rgb_to_owo, rgb_to_owo_on, theme_fg_dimmed};
