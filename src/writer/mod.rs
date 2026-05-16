#[derive(Clone, Copy, Debug)]
#[non_exhaustive]
pub enum AsyncWriterTarget {
    Stdout,
    Stderr,
}

#[cfg(feature = "custom-async")]
pub mod custom;

#[cfg(feature = "custom-async")]
pub use custom::{AsyncWriter, async_writer, async_writer_for};

#[cfg(feature = "native-async")]
pub mod native;

#[cfg(feature = "native-async")]
pub use native::{NativeAsyncWriter, native_async_writer};

#[cfg(feature = "file")]
pub mod file;

pub use file::resolve_log_path;
#[cfg(feature = "file")]
pub use file::{FileWriter, LogHandle, build_file_layer, rotate_log_file};
