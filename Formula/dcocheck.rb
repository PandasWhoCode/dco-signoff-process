class Dcocheck < Formula
  desc "DCO sign-off checker for git repositories"
  homepage "https://github.com/PandasWhoCode/dco-signoff-process"
  version "1.0.0"
  license "Apache-2.0"

  on_macos do
    on_arm do
      url "https://github.com/PandasWhoCode/dco-signoff-process/releases/download/v#{version}/dcocheck-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER_DARWIN_ARM64_SHA256" # darwin-arm64
    end
    on_intel do
      url "https://github.com/PandasWhoCode/dco-signoff-process/releases/download/v#{version}/dcocheck-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER_DARWIN_AMD64_SHA256" # darwin-amd64
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/PandasWhoCode/dco-signoff-process/releases/download/v#{version}/dcocheck-linux-arm64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_ARM64_SHA256" # linux-arm64
    end
    on_intel do
      url "https://github.com/PandasWhoCode/dco-signoff-process/releases/download/v#{version}/dcocheck-linux-amd64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_AMD64_SHA256" # linux-amd64
    end
  end

  def install
    bin.install "dcocheck"
  end

  test do
    assert_match "dcocheck version", shell_output("#{bin}/dcocheck --version")
  end
end
