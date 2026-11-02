#!/usr/bin/env python3
"""Validate release metadata and every archive before allowing publication."""

import argparse
import hashlib
import json
from pathlib import Path
import re
import shlex
import subprocess
import tarfile
import zipfile


VERSION = re.compile(r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)")


def version_tuple(value: str) -> tuple[int, ...]:
    match = VERSION.fullmatch(value)
    if not match:
        raise ValueError(f"Invalid stable version: {value!r}; expected vX.Y.Z")
    return tuple(int(part) for part in match.groups())


def validate_metadata(release: dict, version: str, tags: list[str]) -> str:
    """Reject prereleases, downgrades and releases without a pinned commit."""
    current = version_tuple(version)
    if release["tagName"] != version or release["isPrerelease"]:
        raise ValueError("Release must match the requested version and not be a prerelease")
    newer = [tag for tag in tags if VERSION.fullmatch(tag) and version_tuple(tag) > current]
    if newer:
        raise ValueError(f"Refusing stale release {version}; newer tags exist: {', '.join(newer)}")
    commit = release["targetCommitish"]
    if not re.fullmatch(r"[0-9a-f]{40}", commit):
        raise ValueError("Release target must be the full build commit SHA, not a branch")
    if release["isDraft"] and version in tags:
        raise ValueError("Draft tag already exists remotely; resolve the conflicting tag first")
    if not release["isDraft"] and version not in tags:
        raise ValueError("Published release has no matching tag")
    return commit


def validate_assets(directory: Path, version: str) -> list[Path]:
    """Check all six checksums and extract only the expected regular binaries."""
    version_tuple(version)
    expected = {
        f"qq-{version}-{system}-{arch}.{'zip' if system == 'windows' else 'tar.gz'}"
        for system in ("linux", "darwin", "windows")
        for arch in ("amd64", "arm64")
    }
    checksums: dict[str, str] = {}
    for line in (directory / "checksums.txt").read_text().splitlines():
        match = re.fullmatch(r"([0-9a-f]{64})  (.+)", line)
        if not match or match[2] in checksums:
            raise ValueError("Malformed or duplicate checksum entry")
        checksums[match[2]] = match[1]
    if checksums.keys() != expected:
        raise ValueError("Checksum manifest must contain exactly the six release archives")
    output = directory / "verified"
    output.mkdir(exist_ok=True)
    binaries = []
    for name in sorted(expected):
        archive = directory / name
        if hashlib.sha256(archive.read_bytes()).hexdigest() != checksums[name]:
            raise ValueError(f"Checksum mismatch: {name}")
        binary = output / name
        if name.endswith(".zip"):
            with zipfile.ZipFile(archive) as bundle:
                entries = [entry for entry in bundle.infolist() if entry.filename == "qq.exe"]
                if len(entries) != 1 or entries[0].is_dir():
                    raise ValueError(f"Expected exactly one qq.exe binary in {name}")
                data = bundle.read(entries[0])
        else:
            with tarfile.open(archive) as bundle:
                entries = [entry for entry in bundle.getmembers() if entry.name == "qq"]
                if len(entries) != 1 or not entries[0].isfile():
                    raise ValueError(f"Expected exactly one regular qq binary in {name}")
                stream = bundle.extractfile(entries[0])
                if stream is None:
                    raise ValueError(f"Cannot read binary in {name}")
                data = stream.read()
        binary.write_bytes(data)
        binaries.append(binary)
    return binaries


def validate_build_info(info: str, version: str, commit: str) -> None:
    """Check each platform's embedded version and source revision without executing it."""
    settings = {}
    for line in info.splitlines():
        if line.startswith("\tbuild\t"):
            key, _, value = line.removeprefix("\tbuild\t").partition("=")
            settings[key] = value
    flags = settings.get("-ldflags", "")
    # Go quotes the entire flag string when it contains spaces.
    if flags.startswith('"'):
        flags = json.loads(flags)
    tokens = shlex.split(flags)
    expected = f"github.com/JFryy/qq/cli.Version={version}"
    injected = [tokens[i + 1] for i, token in enumerate(tokens[:-1]) if token == "-X"]
    if expected not in injected:
        raise ValueError(f"Binary does not embed the expected version {version}")
    if settings.get("vcs.revision") != commit or settings.get("vcs.modified") != "false":
        raise ValueError("Binary was not built from the expected clean commit")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("version")
    parser.add_argument("metadata", type=Path)
    parser.add_argument("assets", type=Path)
    parser.add_argument("--metadata-only", action="store_true")
    args = parser.parse_args()
    release = json.loads(args.metadata.read_text())
    tags = subprocess.check_output(["git", "tag", "--list"], text=True).splitlines()
    commit = validate_metadata(release, args.version, tags)
    subprocess.run(["git", "merge-base", "--is-ancestor", commit, "origin/main"], check=True)
    if not release["isDraft"]:
        tagged = subprocess.check_output(["git", "rev-list", "-n1", args.version], text=True).strip()
        if tagged != commit:
            raise ValueError("Published tag does not point to the build commit")
    if args.metadata_only:
        return
    for binary in validate_assets(args.assets, args.version):
        info = subprocess.check_output(["go", "version", "-m", str(binary)], text=True)
        validate_build_info(info, args.version, commit)
    native = args.assets / "verified" / f"qq-{args.version}-linux-amd64.tar.gz"
    native.chmod(0o700)
    actual = subprocess.check_output([str(native), "--version"], text=True).strip()
    if actual != f"qq version {args.version}":
        raise ValueError(f"Binary reports {actual!r}, expected 'qq version {args.version}'")


if __name__ == "__main__":
    try:
        main()
    except (ValueError, OSError, subprocess.CalledProcessError, tarfile.TarError, zipfile.BadZipFile) as error:
        raise SystemExit(f"Release validation failed: {error}") from error
