#!/usr/bin/env python3
"""Create a public source tree without private repository dependencies."""

from __future__ import annotations

import os
import re
import shutil
import subprocess
import time
from pathlib import Path


PRIVATE_MODULES = (
    "github.com/oneclickvirt/privatespeedtest",
    "github.com/oneclickvirt/security",
)
PRIVATE_REFERENCE = re.compile(
    r"github\.com/oneclickvirt/(?:privatespeedtest|security)(?:/|[\"`]|$)",
    flags=re.IGNORECASE,
)
TEXT_SUFFIXES = {".go", ".mod", ".sum", ".md", ".mdx", ".py", ".sh", ".yaml", ".yml"}
IGNORED_DIRECTORIES = {".git", "vendor", ".cache", ".tmp"}


def read_file(path: str) -> str:
    return Path(path).read_text(encoding="utf-8")


def write_file(path: str, content: str) -> None:
    Path(path).write_text(content, encoding="utf-8")


def remove_direct_module_requirements() -> None:
    path = "go.mod"
    content = read_file(path)
    for module in PRIVATE_MODULES:
        content, replacements = re.subn(
            rf"(?m)^\s*{re.escape(module)}\s+\S+(?:\s+//.*)?\r?\n",
            "",
            content,
        )
        if replacements != 1:
            raise ValueError(f"Expected one direct requirement for {module}, found {replacements}")
    write_file(path, content)


def modify_gui_utils() -> None:
    path = "utils/utils.go"
    content = read_file(path)

    content, replacements = re.subn(
        r'(?m)^\s*"github\.com/oneclickvirt/security/network"\r?\n',
        "",
        content,
    )
    if replacements != 1:
        raise ValueError(f"Expected one private network import in {path}, found {replacements}")

    content, replacements = re.subn(
        r"(?<![A-Za-z0-9_])network\.NetworkCheck",
        "bnetwork.NetworkCheck",
        content,
    )
    if replacements != 1:
        raise ValueError(f"Expected one private network check in {path}, found {replacements}")

    content, replacements = re.subn(
        r"\ttoken := network\.SecurityUploadToken",
        r'\ttoken := "OvwKx5qgJtf7PZgCKbtyojSU.MTcwMTUxNzY1MTgwMw"',
        content,
    )
    if replacements != 1:
        raise ValueError(f"Expected one upload-token accessor in {path}, found {replacements}")

    write_file(path, content)


def remove_private_release_contract() -> None:
    path = "main_test.go"
    content = read_file(path)
    content, replacements = re.subn(
        r'(?m)^\s*privatepst "github\.com/oneclickvirt/privatespeedtest/pst"\r?\n',
        "",
        content,
    )
    if replacements != 1:
        raise ValueError(f"Expected one private speedtest test import in {path}, found {replacements}")

    content, replacements = re.subn(
        r'\n\tif got := privatepst\.PrivateSpeedTestVersion; got != "[^"]+" \{'
        r'\n\t\tt\.Fatalf\("private speedtest component version = %q, want [^"]+", got\)'
        r'\n\t\}\n',
        "\n",
        content,
    )
    if replacements != 1:
        raise ValueError(f"Expected one private speedtest release assertion in {path}, found {replacements}")
    write_file(path, content)


def run_go(*args: str, extra_environment: dict[str, str] | None = None) -> None:
    environment = os.environ.copy()
    environment["GOWORK"] = "off"
    if extra_environment:
        environment.update(extra_environment)
    for attempt in range(1, 4):
        completed = subprocess.run(
            ["go", *args],
            check=False,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            env=environment,
        )
        if completed.returncode == 0:
            return
        if attempt < 3:
            time.sleep(attempt * 3)
    raise RuntimeError(f"go {' '.join(args)} failed after 3 attempts")


def use_public_goecs() -> None:
    # Resolve the independently published public branch after local private
    # imports have been removed. A ref rather than a release tag is required:
    # normal release tags intentionally retain the private dependency chain.
    module = "github.com/oneclickvirt/ecs"
    checksum_patterns = [
        pattern
        for pattern in os.environ.get("GONOSUMDB", "").split(",")
        if pattern
    ]
    if module not in checksum_patterns:
        checksum_patterns.append(module)
    # The public branch is force-updated after every release. A module proxy
    # can serve an older branch tip successfully, so always resolve this
    # mutable source ref directly instead of treating a cached result as valid.
    public_environment = {
        "GONOSUMDB": ",".join(checksum_patterns),
        "GOPROXY": "direct",
    }
    run_go("get", f"{module}@public", extra_environment=public_environment)
    run_go("mod", "tidy", extra_environment=public_environment)


def remove_private_delivery_artifacts() -> None:
    """Remove workflows that would load private modules on the public branch."""
    for path in (
        ".back/create_public_branch.py",
        ".github/workflows/build.yml",
        ".github/workflows/build_public.yml",
    ):
        target = Path(path)
        if not target.is_file():
            raise FileNotFoundError(f"Expected delivery artifact is missing: {path}")
        target.unlink()

    vendor = Path("vendor")
    if vendor.is_dir():
        shutil.rmtree(vendor)


def validate_public_tree() -> None:
    matches: list[str] = []
    # Keep documentation and normal application text unchanged. Only Go build
    # inputs and workflows can make a public checkout load a private module.
    for path in (Path("go.mod"), Path("go.sum")):
        if PRIVATE_REFERENCE.search(read_file(str(path))):
            matches.append(str(path))
    for directory, subdirectories, filenames in os.walk("."):
        subdirectories[:] = [name for name in subdirectories if name not in IGNORED_DIRECTORIES]
        for filename in filenames:
            path = Path(directory, filename)
            if path.suffix.lower() != ".go" and path.parts[:2] != (".github", "workflows"):
                continue
            if PRIVATE_REFERENCE.search(read_file(str(path))):
                matches.append(str(path))
    if matches:
        raise ValueError(f"Public tree still references private modules: {', '.join(sorted(matches))}")

    environment = os.environ.copy()
    environment["GOWORK"] = "off"
    completed = subprocess.run(
        ["go", "list", "-deps", "./..."],
        check=False,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        env=environment,
    )
    if completed.returncode != 0:
        raise RuntimeError("could not resolve the public dependency graph")
    loaded_private = [
        line
        for line in completed.stdout.splitlines()
        if any(line == module or line.startswith(module + "/") for module in PRIVATE_MODULES)
    ]
    if loaded_private:
        raise ValueError(f"Public build still loads private modules: {', '.join(loaded_private)}")


def main() -> None:
    if not Path("go.mod").is_file() or not Path("utils/utils.go").is_file():
        raise RuntimeError("Run this script from the ecs-gui repository root")

    remove_direct_module_requirements()
    modify_gui_utils()
    remove_private_release_contract()
    use_public_goecs()
    remove_private_delivery_artifacts()
    # A final tidy runs after private build inputs have gone. Besides keeping
    # module metadata minimal, this removes stale restricted-module checksums
    # that may survive the earlier public-branch switch on newer Go releases.
    run_go("mod", "tidy")
    validate_public_tree()
    print("Public GUI source tree generated successfully")


if __name__ == "__main__":
    main()
