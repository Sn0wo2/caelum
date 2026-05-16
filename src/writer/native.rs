use tracing_subscriber::fmt::MakeWriter;

use super::AsyncWriterTarget;

#[allow(clippy::module_name_repetitions)]
pub struct NativeAsyncWriter {
    writer: tracing_appender::non_blocking::NonBlocking,
    _guard: tracing_appender::non_blocking::WorkerGuard,
}

impl std::fmt::Debug for NativeAsyncWriter {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("NativeAsyncWriter").finish_non_exhaustive()
    }
}

impl MakeWriter<'_> for NativeAsyncWriter {
    type Writer = tracing_appender::non_blocking::NonBlocking;

    fn make_writer(&self) -> Self::Writer {
        self.writer.to_owned()
    }
}

#[allow(clippy::module_name_repetitions)]
pub fn native_async_writer(target: AsyncWriterTarget) -> NativeAsyncWriter {
    let (writer, guard) = match target {
        AsyncWriterTarget::Stdout => tracing_appender::non_blocking(std::io::stdout()),
        AsyncWriterTarget::Stderr => tracing_appender::non_blocking(std::io::stderr()),
    };

    NativeAsyncWriter {
        writer,
        _guard: guard,
    }
}
