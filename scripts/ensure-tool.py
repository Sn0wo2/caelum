#!/usr/bin/env python3
"""Ensure a CLI tool is installed. Usage: ensure-tool.py <tool> -- <install-cmd...>"""

import sys
from common import has_tool, run


def main():
    if len(sys.argv) < 2:
        print("Usage: ensure-tool.py <tool> -- <install-cmd...>", file=sys.stderr)
        sys.exit(1)

    tool = sys.argv[1]
    if has_tool(tool):
        print(f"{tool} already installed")
        sys.exit(0)

    try:
        sep = sys.argv.index("--")
        install_cmd = sys.argv[sep + 1 :]
    except ValueError:
        print(f"{tool} not found and no install command provided", file=sys.stderr)
        sys.exit(1)

    print(f"Installing {tool}...")
    run(install_cmd)


if __name__ == "__main__":
    main()
