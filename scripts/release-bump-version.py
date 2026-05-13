#!/usr/bin/env python3
"""Bump version in Cargo.toml and update workspace dependency."""

import sys
from common import run


def main():
    version = sys.argv[1]

    import tomlkit

    with open("Cargo.toml", encoding="utf-8") as f:
        doc = tomlkit.parse(f.read())

    doc["workspace"]["package"]["version"] = version
    doc["workspace"]["dependencies"]["acta-build"]["version"] = version

    with open("Cargo.toml", "w", encoding="utf-8") as f:
        f.write(tomlkit.dumps(doc))

    run(["cargo", "update", "-w"])
    print(f"Version bumped to {version}")


if __name__ == "__main__":
    main()
