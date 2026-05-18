use rayon::prelude::*;
use walkdir::WalkDir;

pub fn walk_src_max_width(dir: &str, strip_prefix: &str) -> Result<usize, String> {
    let entries: Vec<_> = WalkDir::new(dir)
        .into_iter()
        .filter_map(|e| e.ok())
        .filter(|e| e.path().extension().is_some_and(|ext| ext == "rs"))
        .collect();

    if entries.is_empty() {
        return Err(format!("no .rs files found in {dir}"));
    }

    let max = entries
        .par_iter()
        .map(|e| {
            let display = e.path().to_string_lossy().replace('\\', "/");
            display
                .find(strip_prefix)
                .map(|i| &display[i + strip_prefix.len()..])
                .unwrap_or(&display)
                .len()
        })
        .max()
        .unwrap_or(0)
        + 4;

    Ok(max)
}
