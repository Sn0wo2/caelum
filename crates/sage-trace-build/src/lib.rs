use std::path::Path;

/// Walk a directory recursively and find the longest `.rs` file path.
/// Returns the suggested width for path formatting (longest path + 4 for line numbers).
///
/// # Arguments
/// * `dir` - Directory to scan (e.g., "src")
/// * `strip_prefix` - Prefix to strip from paths (e.g., "src/")
///
/// # Example
/// ```no_run
/// // In your build.rs:
/// fn main() {
///     let max = sage_trace_build::walk_src_max_width("src", "src/");
///     println!("cargo::rustc-env=SAGE_TRACE_MAX_PATH_WIDTH={max}");
/// }
/// ```
pub fn walk_src_max_width(dir: &str, strip_prefix: &str) -> usize {
    let path = Path::new(dir);
    let raw_max = walk_recursive(path, strip_prefix, 0);
    raw_max + 4
}

fn walk_recursive(dir: &Path, prefix: &str, current_max: usize) -> usize {
    let mut max = current_max;
    let Ok(entries) = std::fs::read_dir(dir) else {
        return max;
    };
    for entry in entries.flatten() {
        let path = entry.path();
        if path.is_dir() {
            max = walk_recursive(&path, prefix, max);
        } else if path.extension().is_some_and(|e| e == "rs") {
            let display = path.to_string_lossy().replace('\\', "/");
            let stripped = display
                .find(prefix)
                .map(|i| &display[i + prefix.len()..])
                .unwrap_or(&display);
            if stripped.len() > max {
                max = stripped.len();
            }
        }
    }
    max
}
