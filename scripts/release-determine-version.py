#!/usr/bin/env python3
"""Determine release version and write to GITHUB_ENV."""

import os
import sys
from common import gh_env


def main():
    event = os.environ.get("GITHUB_EVENT_NAME", "")
    if event == "push":
        ref = os.environ.get("GITHUB_REF_NAME", "")
        version = ref.lstrip("v")
    else:
        version = sys.argv[1] if len(sys.argv) > 1 else ""

    if not version:
        print("Could not determine version", file=sys.stderr)
        sys.exit(1)

    gh_env("VERSION", version)
    print(f"VERSION={version}")


if __name__ == "__main__":
    main()
