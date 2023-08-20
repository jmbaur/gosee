{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorSha256 = "sha256-B0dGWPSdl5s6mYr4w5ApcDvW2J9gxBHqaym0r+sYXU4=";
  ldflags = [ "-s" "-w" ];
}
