#!/bin/bash

# Script to update Homebrew formula for EnvTrack CLI

set -e

# Check if all required arguments are provided
if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <version> <github_token> <tap_repo>"
    exit 1
fi

VERSION=$1
GITHUB_TOKEN=$2
TAP_REPO=$3

# Function to calculate SHA256
calculate_sha256() {
    local file=$1
    if [[ "$OSTYPE" == "darwin"* ]]; then
        shasum -a 256 "$file" | awk '{print $1}'
    else
        sha256sum "$file" | awk '{print $1}'
    fi
}

# Clone the Homebrew tap repository
git clone https://x-access-token:${GITHUB_TOKEN}@github.com/${TAP_REPO}.git homebrew-tap
cd homebrew-tap

# Create Formula directory if it doesn't exist
mkdir -p Formula

# Generate the formula file
cat > Formula/envtrack.rb << EOL
class Envtrack < Formula
  desc "EnvTrack CLI tool for managing environment variables"
  homepage "https://github.com/${GITHUB_REPOSITORY}"
  version "${VERSION}"
  license "MIT"

  if OS.mac?
    if Hardware::CPU.arm?
      url "https://github.com/${GITHUB_REPOSITORY}/releases/download/${VERSION}/envtrack-${VERSION}-darwin-arm64"
      sha256 "$(calculate_sha256 ../dist/envtrack-${VERSION}-darwin-arm64)"
    else
      url "https://github.com/${GITHUB_REPOSITORY}/releases/download/${VERSION}/envtrack-${VERSION}-darwin-amd64"
      sha256 "$(calculate_sha256 ../dist/envtrack-${VERSION}-darwin-amd64)"
    end
  elsif OS.linux?
    if Hardware::CPU.arm?
      url "https://github.com/${GITHUB_REPOSITORY}/releases/download/${VERSION}/envtrack-${VERSION}-linux-arm64"
      sha256 "$(calculate_sha256 ../dist/envtrack-${VERSION}-linux-arm64)"
    else
      url "https://github.com/${GITHUB_REPOSITORY}/releases/download/${VERSION}/envtrack-${VERSION}-linux-amd64"
      sha256 "$(calculate_sha256 ../dist/envtrack-${VERSION}-linux-amd64)"
    end
  end

  def install
    bin.install "envtrack"
  end

  test do
    assert_match "EnvTrack CLI version ${VERSION}", shell_output("#{bin}/envtrack version")
  end
end
EOL

# Commit and push changes
git config user.name "GitHub Actions Bot"
git config user.email "<>"
git add Formula/envtrack.rb
git commit -m "Update EnvTrack formula to version ${VERSION}"
git push

echo "Homebrew formula updated successfully!"