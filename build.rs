fn main() {
    let max = caelum_build::walk_src_max_width("src", "src/");
    std::fs::write(
        std::path::Path::new(&std::env::var("OUT_DIR").expect("Cargo should set OUT_DIR"))
            .join("path_width"),
        max.to_string(),
    )
    .expect("failed to write path_width to OUT_DIR");
    println!("cargo::rerun-if-changed=src");
    println!("cargo::warning=Max Path Width: {max}");
}
