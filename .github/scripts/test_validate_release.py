"""Regression tests for release publication gates; no network or remote changes."""

import hashlib
import importlib.util
import io
from pathlib import Path
import tarfile
import tempfile
import unittest
import zipfile

spec = importlib.util.spec_from_file_location("validate_release", Path(__file__).with_name("validate-release.py"))
release = importlib.util.module_from_spec(spec)
spec.loader.exec_module(release)


class ReleaseValidationTests(unittest.TestCase):
    def metadata(self, **changes):
        return dict(tagName="v0.3.5", isDraft=True, isPrerelease=False,
                    targetCommitish="a" * 40, **changes)

    def test_valid_draft_and_published_retry(self):
        metadata = self.metadata()
        self.assertEqual(release.validate_metadata(metadata, "v0.3.5", ["v0.3.4"]), "a" * 40)
        metadata["isDraft"] = False
        release.validate_metadata(metadata, "v0.3.5", ["v0.3.4", "v0.3.5"])

    def test_invalid_versions(self):
        for value in ["v01.2.3", "1.2.3", "v1.2.3-rc1", "v1.2", "v1.2.3\n"]:
            with self.subTest(value=value), self.assertRaises(ValueError):
                release.version_tuple(value)

    def test_stale_release_is_rejected_for_draft_and_retry(self):
        for draft in [True, False]:
            metadata = self.metadata()
            metadata["isDraft"] = draft
            with self.assertRaisesRegex(ValueError, "stale"):
                release.validate_metadata(metadata, "v0.3.5", ["v0.3.5", "v0.10.0"])

    def test_prerelease_wrong_tag_and_unpinned_commit_rejected(self):
        for key, value in [("isPrerelease", True), ("tagName", "v0.3.6"), ("targetCommitish", "main")]:
            metadata = self.metadata()
            metadata[key] = value
            with self.subTest(key=key), self.assertRaises(ValueError):
                release.validate_metadata(metadata, "v0.3.5", ["v0.3.4"])

    def test_draft_cannot_reuse_existing_tag(self):
        with self.assertRaisesRegex(ValueError, "already exists"):
            release.validate_metadata(self.metadata(), "v0.3.5", ["v0.3.5"])

    def test_build_info_checks_version_revision_and_clean_tree(self):
        info = ('\tbuild\t-ldflags="-s -w -X github.com/JFryy/qq/cli.Version=v0.3.5"\n'
                f'\tbuild\tvcs.revision={"a" * 40}\n\tbuild\tvcs.modified=false\n')
        release.validate_build_info(info, "v0.3.5", "a" * 40)
        for bad in [info.replace("v0.3.5", "v0.3.4"), info.replace("false", "true"),
                    info.replace("a" * 40, "b" * 40), "not a Go binary"]:
            with self.subTest(info=bad), self.assertRaises(ValueError):
                release.validate_build_info(bad, "v0.3.5", "a" * 40)

    def assets(self, directory):
        checksums = []
        for system in ["linux", "darwin", "windows"]:
            for arch in ["amd64", "arm64"]:
                suffix = "zip" if system == "windows" else "tar.gz"
                archive = directory / f"qq-v0.3.5-{system}-{arch}.{suffix}"
                if system == "windows":
                    with zipfile.ZipFile(archive, "w") as bundle:
                        bundle.writestr("qq.exe", b"binary")
                else:
                    with tarfile.open(archive, "w:gz") as bundle:
                        entry = tarfile.TarInfo("qq")
                        entry.size = 6
                        bundle.addfile(entry, io.BytesIO(b"binary"))
                checksums.append(f"{hashlib.sha256(archive.read_bytes()).hexdigest()}  {archive.name}\n")
        (directory / "checksums.txt").write_text("".join(checksums))

    def test_all_six_assets_are_required_and_verified(self):
        with tempfile.TemporaryDirectory() as temp:
            directory = Path(temp)
            self.assets(directory)
            self.assertEqual(len(release.validate_assets(directory, "v0.3.5")), 6)
            archive = directory / "qq-v0.3.5-windows-arm64.zip"
            archive.unlink()
            with self.assertRaises(FileNotFoundError):
                release.validate_assets(directory, "v0.3.5")
            self.assets(directory)
            archive.write_bytes(b"corrupt")
            with self.assertRaisesRegex(ValueError, "Checksum mismatch"):
                release.validate_assets(directory, "v0.3.5")

    def test_missing_and_duplicate_checksum_entries_rejected(self):
        with tempfile.TemporaryDirectory() as temp:
            directory = Path(temp)
            self.assets(directory)
            manifest = directory / "checksums.txt"
            lines = manifest.read_text().splitlines(keepends=True)
            for content in ["".join(lines[1:]), "".join(lines + lines[:1])]:
                manifest.write_text(content)
                with self.assertRaises(ValueError):
                    release.validate_assets(directory, "v0.3.5")

    def test_malformed_archive_with_valid_checksum_rejected(self):
        with tempfile.TemporaryDirectory() as temp:
            directory = Path(temp)
            self.assets(directory)
            archive = directory / "qq-v0.3.5-linux-amd64.tar.gz"
            original = hashlib.sha256(archive.read_bytes()).hexdigest()
            with tarfile.open(archive, "w:gz") as bundle:
                entry = tarfile.TarInfo("qq")
                entry.type = tarfile.SYMTYPE
                entry.linkname = "/bin/true"
                bundle.addfile(entry)
            manifest = directory / "checksums.txt"
            manifest.write_text(manifest.read_text().replace(original, hashlib.sha256(archive.read_bytes()).hexdigest()))
            with self.assertRaisesRegex(ValueError, "regular qq"):
                release.validate_assets(directory, "v0.3.5")


if __name__ == "__main__":
    unittest.main()
