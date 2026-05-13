#!/usr/bin/env python3
"""Check if git working tree has changes, write result to GITHUB_OUTPUT."""

from common import run_ok, gh_output


def main():
    result = run_ok(["git", "diff", "--quiet"])
    has_changes = "false" if result.returncode == 0 else "true"
    gh_output("has_changes", has_changes)
    print(f"has_changes={has_changes}")


if __name__ == "__main__":
    main()
