#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-}"
SOURCE_SHA256="${2:-}"
OUTPUT="${3:-$ROOT/dist/tasklight.rb}"
REPOSITORY="${TASKLIGHT_REPOSITORY:-revazi/tasklight}"

if [[ ! "$VERSION" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
  echo "usage: $0 <X.Y.Z> <source-sha256> [output.rb]" >&2
  exit 2
fi
if [[ ! "$SOURCE_SHA256" =~ ^[a-f0-9]{64}$ ]]; then
  echo "source SHA-256 must contain 64 lowercase hexadecimal characters" >&2
  exit 2
fi

mkdir -p "$(dirname "$OUTPUT")"
cat >"$OUTPUT" <<RUBY
class Tasklight < Formula
  desc "Terminal task runner with desktop notifications"
  homepage "https://github.com/$REPOSITORY"
  url "https://github.com/$REPOSITORY/releases/download/v$VERSION/tasklight-v$VERSION-source.tar.gz"
  sha256 "$SOURCE_SHA256"
  license "MIT"
  head "https://github.com/$REPOSITORY.git", branch: "main"

  depends_on "go" => :build

  on_macos do
    depends_on xcode: ["13.0", :build]
  end

  def install
    ldflags = "-s -w -X github.com/revazi/tasklight/internal/cli.Version=#{version}"
    system "go", "build", "-trimpath", "-ldflags", ldflags, "-o", "tasklight", "./cmd/tasklight"

    unless OS.mac?
      bin.install "tasklight"
      return
    end

    target = Hardware::CPU.arm? ? "darwin-arm64" : "darwin-amd64"
    helper_dir = buildpath/"homebrew-helper"
    helper_version = version.to_s.match?(/\A\d+\.\d+\.\d+\z/) ? version.to_s : "0.0.0"
    ENV["TASKLIGHT_VERSION"] = helper_version
    ENV["TASKLIGHT_SKIP_REGISTER"] = "1"
    system "./scripts/build-macos-helper.sh", helper_dir, target
    libexec.install "tasklight", helper_dir/"Tasklight.app"
    (bin/"tasklight").write_env_script libexec/"tasklight", TASKLIGHT_MACOS_HELPER: libexec/"Tasklight.app"
  end

  test do
    assert_match "tasklight #{version}", shell_output("#{bin}/tasklight --version")
    assert_match "Usage:", shell_output("#{bin}/tasklight --help")
  end
end
RUBY

printf 'generated Homebrew formula %s\n' "$OUTPUT"
