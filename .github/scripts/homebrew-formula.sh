#!/usr/bin/env bash
# Renders the Homebrew formula for a qq release from its checksums.txt.
# Usage: homebrew-formula.sh <version> <checksums.txt>
set -euo pipefail

if [ $# -ne 2 ]; then
  echo "usage: $0 <version> <checksums.txt>" >&2
  exit 2
fi

version="$1"
checksums="$2"
base="https://github.com/JFryy/qq/releases/download/${version}"

sha() {
  local file="qq-${version}-$1.tar.gz"
  local sum
  sum="$(awk -v f="$file" '$2 == f { print $1 }' "$checksums")"
  if [ -z "$sum" ]; then
    echo "error: no checksum for $file in $checksums" >&2
    exit 1
  fi
  echo "$sum"
}

darwin_amd64="$(sha darwin-amd64)"
darwin_arm64="$(sha darwin-arm64)"
linux_amd64="$(sha linux-amd64)"
linux_arm64="$(sha linux-arm64)"

cat <<EOF
class Qq < Formula
  desc "Multi-tool structured format processor for query and transcoding"
  homepage "https://github.com/JFryy/qq"
  version "${version#v}"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "${base}/qq-${version}-darwin-amd64.tar.gz"
      sha256 "${darwin_amd64}"
    elsif Hardware::CPU.arm?
      url "${base}/qq-${version}-darwin-arm64.tar.gz"
      sha256 "${darwin_arm64}"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "${base}/qq-${version}-linux-amd64.tar.gz"
      sha256 "${linux_amd64}"
    elsif Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "${base}/qq-${version}-linux-arm64.tar.gz"
      sha256 "${linux_arm64}"
    end
  end

  def install
    bin.install "qq"
  end

  test do
    (testpath/"test.json").write('{"somekey": "somevalue"}')
    assert_equal "somevalue", shell_output("cat test.json | #{bin}/qq .somekey -r").strip
    assert_match version.to_s, shell_output("#{bin}/qq --version")
  end
end
EOF
