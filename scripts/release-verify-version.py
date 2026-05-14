#!/usr/bin/env python3
import sys
import tomllib


def main():
    tag_version = sys.argv[1]

    with open("Cargo.toml", "rb") as f:
        doc = tomllib.load(f)

    cargo_version = str(doc["workspace"]["package"]["version"])

    if tag_version != cargo_version:
        print(
            f"Version mismatch: tag v{tag_version} but Cargo.toml has {cargo_version}",
            file=sys.stderr,
        )
        sys.exit(1)

    print(f"Version verified: v{tag_version} matches Cargo.toml")


if __name__ == "__main__":
    main()
