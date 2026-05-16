#!/usr/bin/env python3
"""Download and install macOS SDK for darwin cross-compilation via zigbuild."""

import os
import subprocess
import sys
import tarfile
import urllib.request
from pathlib import Path


SDK_URL = "https://github.com/phracker/MacOSX-SDKs/releases/download/11.3/MacOSX11.3.sdk.tar.xz"
SDK_ROOT = "/opt/MacOSX11.3.sdk"


def main():
    if Path(SDK_ROOT).is_dir():
        print(f"SDK already installed at {SDK_ROOT}")
        _export_sdkroot()
        return

    print(f"Downloading macOS SDK from {SDK_URL}")
    tmp_path = Path("/tmp/MacOSX11.3.sdk.tar.xz")
    urllib.request.urlretrieve(SDK_URL, tmp_path)

    print("Extracting macOS SDK")
    Path("/opt").mkdir(parents=True, exist_ok=True)
    with tarfile.open(tmp_path, mode="r:xz") as tar:
        tar.extractall(path="/opt", filter="fully_trusted")

    tmp_path.unlink(missing_ok=True)

    if not Path(SDK_ROOT).is_dir():
        print(f"ERROR: SDK extraction failed, {SDK_ROOT} not found", file=sys.stderr)
        sys.exit(1)

    print(f"macOS SDK installed at {SDK_ROOT}")
    _export_sdkroot()


def _export_sdkroot():
    env_file = os.environ.get("GITHUB_ENV")
    if env_file:
        with open(env_file, "a") as f:
            f.write(f"SDKROOT={SDK_ROOT}\n")
        print(f"SDKROOT={SDK_ROOT} written to GITHUB_ENV")


if __name__ == "__main__":
    main()
