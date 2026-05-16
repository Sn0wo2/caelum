#[cfg(feature = "file")]
use std::path::{Path, PathBuf};

use crate::Result;
use crate::reload::ReloadHandle;
#[cfg(feature = "file")]
use crate::writer::LogHandle;

#[must_use = "dropping TracingGuard will release associated resources"]
#[derive(Debug)]
pub struct TracingGuard {
    #[cfg(feature = "file")]
    #[allow(dead_code)]
    pub(crate) worker_guard: Option<LogHandle>,
    #[cfg(feature = "file")]
    pub(crate) log_path: Option<PathBuf>,
    pub(crate) reload_handle: ReloadHandle,
}

impl TracingGuard {
    pub const fn reload_handle(&self) -> &ReloadHandle {
        &self.reload_handle
    }

    pub const fn reload_handle_mut(&mut self) -> &mut ReloadHandle {
        &mut self.reload_handle
    }

    #[cfg(feature = "file")]
    pub fn log_path(&self) -> Option<&Path> {
        self.log_path.as_deref()
    }

    pub fn shutdown(&self) -> Result<()> {
        self.reload_handle.shutdown()
    }
}
