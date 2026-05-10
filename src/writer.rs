#[cfg(feature = "custom-async")]
use std::io::Write;
use std::sync::Arc;
#[cfg(feature = "custom-async")]
use std::sync::atomic::{AtomicUsize, Ordering};

#[cfg(feature = "custom-async")]
use tokio::io::AsyncWrite;
#[cfg(feature = "custom-async")]
use tokio::io::{AsyncWriteExt, stderr, stdout};
#[cfg(feature = "custom-async")]
use tokio::sync::mpsc;
use tracing_subscriber::fmt::MakeWriter;

#[derive(Clone, Copy, Debug)]
pub enum AsyncWriterTarget {
    Stdout,
    Stderr,
}

#[cfg(feature = "custom-async")]
#[derive(Clone, Debug)]
pub struct AsyncWriter {
    sender: mpsc::UnboundedSender<Vec<u8>>,
    count: Arc<AtomicUsize>,
}

#[cfg(feature = "custom-async")]
impl AsyncWriter {
    pub fn pending_writes(&self) -> usize {
        self.count.load(Ordering::Relaxed)
    }
}

#[cfg(feature = "custom-async")]
impl Write for AsyncWriter {
    fn write(&mut self, buf: &[u8]) -> std::io::Result<usize> {
        self.count.fetch_add(1, Ordering::Relaxed);
        self.sender.send(buf.to_vec()).map_err(|_| {
            self.count.fetch_sub(1, Ordering::Relaxed);
            std::io::Error::new(std::io::ErrorKind::BrokenPipe, "async writer closed")
        })?;
        Ok(buf.len())
    }

    fn flush(&mut self) -> std::io::Result<()> {
        Ok(())
    }
}

#[cfg(feature = "custom-async")]
impl MakeWriter<'_> for AsyncWriter {
    type Writer = Self;

    fn make_writer(&self) -> Self::Writer {
        self.clone()
    }
}

#[cfg(feature = "custom-async")]
pub fn async_writer() -> AsyncWriter {
    async_writer_for(AsyncWriterTarget::Stdout)
}

#[cfg(feature = "custom-async")]
pub fn async_writer_for(target: AsyncWriterTarget) -> AsyncWriter {
    let (sender, mut receiver) = mpsc::unbounded_channel::<Vec<u8>>();
    let count = Arc::new(AtomicUsize::new(0));
    let count_clone = count.clone();

    tokio::spawn(async move {
        let writer: &mut (dyn AsyncWrite + Unpin + Send) = match target {
            AsyncWriterTarget::Stdout => &mut stdout(),
            AsyncWriterTarget::Stderr => &mut stderr(),
        };

        while let Some(data) = receiver.recv().await {
            if let Err(e) = writer.write_all(&data).await {
                let _ = writeln!(std::io::stderr(), "async writer error: {e}");
            }
            count_clone.fetch_sub(1, Ordering::Relaxed);
        }
    });

    AsyncWriter { sender, count }
}

#[cfg(feature = "native-async")]
pub(crate) struct NativeAsyncWriter {
    writer: tracing_appender::non_blocking::NonBlocking,
    _guard: tracing_appender::non_blocking::WorkerGuard,
}

#[cfg(feature = "native-async")]
impl std::fmt::Debug for NativeAsyncWriter {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("NativeAsyncWriter").finish_non_exhaustive()
    }
}

#[cfg(feature = "native-async")]
impl MakeWriter<'_> for NativeAsyncWriter {
    type Writer = tracing_appender::non_blocking::NonBlocking;

    fn make_writer(&self) -> Self::Writer {
        self.writer.clone()
    }
}

#[cfg(feature = "native-async")]
pub(crate) fn native_async_writer(target: AsyncWriterTarget) -> NativeAsyncWriter {
    let (writer, guard) = match target {
        AsyncWriterTarget::Stdout => tracing_appender::non_blocking(std::io::stdout()),
        AsyncWriterTarget::Stderr => tracing_appender::non_blocking(std::io::stderr()),
    };

    NativeAsyncWriter {
        writer,
        _guard: guard,
    }
}
