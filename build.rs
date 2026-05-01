fn main() {
    let max = sage_trace_build::walk_src_max_width("src", "src/");
    println!("cargo::rustc-env=SAGE_TRACE_MAX_PATH_WIDTH={max}");
    println!("cargo::rerun-if-changed=src");
}
