#!/usr/bin/env python3
"""Git operations: commit, tag, push for release."""

import os
import sys
from common import run, run_ok


def commit_and_tag(version):
    run(["git", "config", "user.name", "github-actions[bot]"])
    run(["git", "config", "user.email", "github-actions[bot]@users.noreply.github.com"])

    r = run_ok(["git", "status", "--porcelain"], capture_output=True, text=True)
    if r.stdout.strip():
        run(["git", "commit", "-am", f"chore: release v{version}"])

    if os.environ.get("GITHUB_EVENT_NAME", "") != "push":
        run(["git", "tag", f"v{version}"])


def push(version):
    run(["git", "push", "origin", f"v{version}"])

    if os.environ.get("GITHUB_EVENT_NAME", "") == "workflow_dispatch":
        branch = os.environ.get("GITHUB_REPOSITORY_DEFAULT_BRANCH", "main")
        run(["git", "push", "origin", f"HEAD:refs/heads/{branch}"])


def main():
    cmd, version = sys.argv[1], sys.argv[2]
    {"commit-and-tag": commit_and_tag, "push": push}[cmd](version)


if __name__ == "__main__":
    main()
