#!/usr/bin/env python3
"""Run cargo-geiger. Never fails."""

import subprocess


with open("geiger-report.md", "w") as f:
    subprocess.run(
        ["cargo", "+nightly", "geiger", "--all-features", "--output-format", "GitHubMarkdown"],
        stdout=f,
        stderr=subprocess.STDOUT,
    )
