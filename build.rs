fn main() {
    let max = sage_trace_build::walk_src_max_width("src", "src/");
    std::fs::write(
        std::path::Path::new(&std::env::var("OUT_DIR").unwrap()).join("path_width"),
        max.to_string(),
    )
    .unwrap();
    println!("cargo::warning=Max Path Width: {max}");
    println!("cargo::rerun-if-changed=src");
}
