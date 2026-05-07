#!/usr/bin/env python3
"""Auto-detect: use cargo for native builds, cross for cross-compilation."""

import os
import re
import subprocess
import sys


def get_host_target():
    result = subprocess.run(
        ["rustc", "-vV"], capture_output=True, text=True, check=True
    )
    match = re.search(r"host:\s+(\S+)", result.stdout)
    if not match:
        raise RuntimeError("Could not determine host target from rustc -vV")
    return match.group(1)


def get_target_from_args(args):
    for i, arg in enumerate(args):
        if arg == "--target" and i + 1 < len(args):
            return args[i + 1]
        if arg.startswith("--target="):
            return arg.split("=", 1)[1]
    return None


def main():
    args = sys.argv[1:]
    host = get_host_target()
    target = get_target_from_args(args)

    tool = "cargo" if (target is None or target == host) else "cross"

    sys.exit(subprocess.call([tool] + args))


if __name__ == "__main__":
    main()
