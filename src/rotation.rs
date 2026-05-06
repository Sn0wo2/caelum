use std::path::Path;

#[cfg(feature = "compress")]
use std::io::Write;

use crate::config::LogRotation;

pub fn rotate_log_file(path: &Path, mode: LogRotation) -> crate::error::Result<()> {
    if !path.exists() {
        return Ok(());
    }

    match mode {
        LogRotation::None => Ok(()),
        LogRotation::Rename => {
            let renamed = path.with_extension(format!("{}.log", now_timestamp()));
            std::fs::rename(path, renamed)?;
            Ok(())
        }
        #[cfg(feature = "compress")]
        LogRotation::Compress => {
            let gz_path = path.with_extension(format!("{}.log.gz", now_timestamp()));
            let input = std::fs::read(path)?;
            let output = std::fs::File::create(&gz_path)?;
            let mut encoder = flate2::write::GzEncoder::new(output, flate2::Compression::default());
            encoder.write_all(&input)?;
            encoder.finish()?;
            std::fs::remove_file(path)?;
            Ok(())
        }
    }
}

fn now_timestamp() -> String {
    chrono::Local::now().format("%Y-%m-%d_%H-%M-%S").to_string()
}
