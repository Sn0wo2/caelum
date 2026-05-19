use std::io::Write;
use std::sync::Arc;
use std::sync::atomic::{AtomicUsize, Ordering};
use tokio::io::{AsyncWrite, AsyncWriteExt, stderr, stdout};
use tokio::sync::mpsc;
use tracing_subscriber::fmt::MakeWriter;

use super::AsyncWriterTarget;

#[derive(Clone, Debug)]
pub struct AsyncWriter {
    sender: mpsc::Sender<Vec<u8>>,
    count: Arc<AtomicUsize>,
}

impl AsyncWriter {
    pub fn pending_writes(&self) -> usize {
        self.count.load(Ordering::Relaxed)
    }
}

impl Write for AsyncWriter {
    fn write(&mut self, buf: &[u8]) -> std::io::Result<usize> {
        self.count.fetch_add(1, Ordering::Relaxed);
        match self.sender.try_send(buf.to_vec()) {
            Ok(_) => Ok(buf.len()),
            Err(mpsc::error::TrySendError::Full(_)) => {
                self.count.fetch_sub(1, Ordering::Relaxed);
                Ok(buf.len())
            }
            Err(mpsc::error::TrySendError::Closed(_)) => {
                self.count.fetch_sub(1, Ordering::Relaxed);
                Err(std::io::Error::new(std::io::ErrorKind::BrokenPipe, "async writer closed"))
            }
        }
    }

    fn flush(&mut self) -> std::io::Result<()> {
        Ok(())
    }
}

impl MakeWriter<'_> for AsyncWriter {
    type Writer = Self;

    fn make_writer(&self) -> Self::Writer {
        self.clone()
    }
}

pub fn async_writer() -> AsyncWriter {
    async_writer_for(AsyncWriterTarget::Stdout)
}

pub fn async_writer_for(target: AsyncWriterTarget) -> AsyncWriter {
    let (sender, mut receiver) = mpsc::channel::<Vec<u8>>(4096);
    let count = Arc::new(AtomicUsize::new(0));
    let count_clone = count.clone();

    tokio::spawn(async move {
            let writer: &mut (dyn AsyncWrite + Unpin + Send) = match target {
                AsyncWriterTarget::Stdout => &mut stdout(),
                AsyncWriterTarget::Stderr => &mut stderr(),
            };

            while let Some(data) = receiver.recv().await {
                if let Err(e) = writer.write_all(&data).await {
                    let _unused = writeln!(std::io::stderr(), "async writer error: {e}");
                }
                count_clone.fetch_sub(1, Ordering::Relaxed);
            }
    });

    AsyncWriter { sender, count }
}
