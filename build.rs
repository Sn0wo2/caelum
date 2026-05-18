#![allow(clippy::expect_used)]

fn main() {
    let max = match acta_build::walk_src_max_width("src", "src/") {
        Ok(m) => m,
        Err(e) => {
            println!("cargo::warning=walk_src_max_width failed: {e}");
            20
        }
    };
    std::fs::write(
        std::path::Path::new(&std::env::var("OUT_DIR").expect("Cargo should set OUT_DIR"))
            .join("path_width"),
        max.to_string(),
    )
    .expect("failed to write path_width to OUT_DIR");

    println!("cargo::rerun-if-changed=src");
}
