use std::path::{Path, PathBuf};

use tracing_subscriber::fmt::MakeWriter;

use crate::Result;
use crate::config::{File, Rotation};

pub type LogHandle = tracing_appender::non_blocking::WorkerGuard;

#[derive(Clone, Debug)]
#[allow(clippy::module_name_repetitions)]
pub struct FileWriter {
    writer: tracing_appender::non_blocking::NonBlocking,
}

impl MakeWriter<'_> for FileWriter {
    type Writer = tracing_appender::non_blocking::NonBlocking;

    fn make_writer(&self) -> Self::Writer {
        self.writer.clone()
    }
}

impl FileWriter {
    pub const fn new(writer: tracing_appender::non_blocking::NonBlocking) -> Self {
        Self { writer }
    }
}

pub fn build_file_layer(file_config: &File) -> Result<(FileWriter, LogHandle, PathBuf)> {
    let path = &file_config.path;
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)?;
    }

    rotate_log_file(path, file_config.rotation)?;
    let path = resolve_log_path(path);

    let (non_blocking, guard) = tracing_appender::non_blocking(tracing_appender::rolling::never(
        path.parent().unwrap_or(Path::new(".")),
        path.file_name().unwrap_or_default(),
    ));

    Ok((FileWriter::new(non_blocking), guard, path))
}

#[allow(clippy::module_name_repetitions)]
pub fn rotate_log_file(path: &Path, mode: Rotation) -> Result<()> {
    if !path.exists() {
        return Ok(());
    }
    match mode {
        Rotation::None => Ok(()),
        Rotation::Rename => {
            let renamed = path.with_extension(format!("{}.log", now_timestamp()));
            std::fs::rename(path, renamed)?;
            Ok(())
        }
        #[cfg(feature = "compress")]
        Rotation::Compress => {
            use flate2::Compression;
            use flate2::read::GzEncoder;
            use std::io::{BufWriter, Read, Write};

            let gz_path = path.with_extension(format!("{}.log.gz", now_timestamp()));
            let mut input = std::fs::File::open(path)?;
            let output = std::fs::File::create(&gz_path)?;
            let mut encoder = GzEncoder::new(&mut input, Compression::default());
            let mut buf_writer = BufWriter::new(output);
            let mut buf = [0_u8; 64 * 1024];

            loop {
                let n = encoder.read(&mut buf)?;
                if n == 0 {
                    break;
                }
                buf_writer.write_all(buf.get(..n).unwrap_or(&[]))?;
            }
            buf_writer.flush()?;
            drop(buf_writer);

            std::fs::remove_file(path)?;
            Ok(())
        }
    }
}

fn now_timestamp() -> String {
    chrono::Local::now().format("%Y-%m-%d_%H-%M-%S").to_string()
}

pub fn resolve_log_path(path: &Path) -> PathBuf {
    match std::fs::OpenOptions::new()
        .create(true)
        .append(true)
        .open(path)
    {
        Ok(_) => path.to_path_buf(),
        Err(_) => {
            let pid = std::process::id();
            let stem = path
                .file_stem()
                .and_then(|s| s.to_str())
                .unwrap_or("latest");
            let ext = path.extension().and_then(|s| s.to_str()).unwrap_or("log");
            path.with_file_name(format!("{stem}-{pid}.{ext}"))
        }
    }
}
