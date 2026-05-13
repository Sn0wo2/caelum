"""Shared utilities for acta CI scripts — thin wrappers around subprocess + GitHub Actions env."""

import os
import shutil
import subprocess
import sys


def run(cmd, **kwargs):
    """Run command, print it, exit on failure."""
    if isinstance(cmd, str):
        print(f"+ {cmd}")
    else:
        print(f"+ {' '.join(cmd)}")
    r = subprocess.run(cmd, **kwargs)
    if r.returncode != 0:
        sys.exit(r.returncode)
    return r


def run_ok(cmd, **kwargs):
    """Run command that may fail. Returns CompletedProcess."""
    if isinstance(cmd, str):
        print(f"+ {cmd}")
    else:
        print(f"+ {' '.join(cmd)}")
    return subprocess.run(cmd, **kwargs)


def has_tool(name):
    return shutil.which(name) is not None


def gh_output(key, value):
    path = os.environ.get("GITHUB_OUTPUT", "")
    if path:
        with open(path, "a") as f:
            f.write(f"{key}={value}\n")


def gh_env(key, value):
    path = os.environ.get("GITHUB_ENV", "")
    if path:
        with open(path, "a") as f:
            f.write(f"{key}={value}\n")


def gh_summary(text):
    path = os.environ.get("GITHUB_STEP_SUMMARY", "")
    if path:
        with open(path, "a") as f:
            f.write(text)
