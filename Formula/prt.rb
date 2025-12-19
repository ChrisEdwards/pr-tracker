# frozen_string_literal: true

# Homebrew formula for PRT - GitHub PR Tracker
# Repository: https://github.com/ChrisEdwards/prt
#
# Installation:
#   brew tap ChrisEdwards/tap
#   brew install prt
#
class Prt < Formula
  desc "GitHub PR Tracker - Aggregate PR status across multiple repositories"
  homepage "https://github.com/ChrisEdwards/prt"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/ChrisEdwards/prt/releases/download/v#{version}/prt-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER_ARM64_SHA256"
    else
      url "https://github.com/ChrisEdwards/prt/releases/download/v#{version}/prt-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER_AMD64_SHA256"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/ChrisEdwards/prt/releases/download/v#{version}/prt-linux-amd64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_AMD64_SHA256"
    end
  end

  depends_on "gh" => :recommended

  def install
    bin.install "prt"
  end

  def caveats
    <<~EOS
      PRT requires the GitHub CLI (gh) to be installed and authenticated.

      If not already installed:
        brew install gh

      Then authenticate:
        gh auth login

      Run `prt` to start the setup wizard.

      Configuration file: ~/.prt/config.yaml
    EOS
  end

  test do
    assert_match "prt version", shell_output("#{bin}/prt --version")
  end
end
