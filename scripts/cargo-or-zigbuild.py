#!/usr/bin/env python3
"""Auto-detect: use cargo for native builds, cargo-zigbuild for cross-compilation."""

import re
import subprocess
import sys


def get_host_target():
    result = subprocess.run(
        ["rustc", "-vV"], capture_output=True, text=True, check=True
    )
    if match := re.search(r"host:\s+(\S+)", result.stdout):
        return match[1]
    else:
        raise RuntimeError("Could not determine host target from rustc -vV")


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

    if target is None or target == host:
        command = ["cargo"]
    else:
        command = ["cargo", "zigbuild"]
        if args and args[0] == "build":
            args = args[1:]

    sys.exit(subprocess.call(command + args))


if __name__ == "__main__":
    main()
