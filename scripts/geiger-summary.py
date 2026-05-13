#!/usr/bin/env python3
"""Append geiger report tail to GITHUB_STEP_SUMMARY."""

from common import gh_summary


try:
    with open("geiger-report.md", encoding="utf-8") as f:
        lines = f.readlines()
    tail = "".join(lines[-50:])
except FileNotFoundError:
    tail = "geiger-report.md not found"

gh_summary(f"## Cargo Geiger Report\n```\n{tail}```\n")
