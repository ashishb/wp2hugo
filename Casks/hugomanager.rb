cask "hugomanager" do
  arch arm: "arm64", intel: "x86_64"

  version "1.3.0"
  sha256 arm:   "5640daed0ed0556a097d4bedda6e20f18fbcd49ef9a16fb2d858e2c6e0f9f552",
         intel: "b91bd2e3b6443a96385ee89b6af503cf7711ca9a35b24cd65f88b84e55f795e5"

  url "https://github.com/ashishb/wp2hugo/releases/download/#{version}-hugomanager/wp2hugo_Darwin_#{arch}.tar.gz"
  name "hugomanager"
  desc "Tool for managing Hugo sites"
  homepage "https://github.com/ashishb/wp2hugo"

  depends_on :macos

  binary "#{staged_path}/hugomanager"
end
