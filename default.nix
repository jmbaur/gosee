{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorSha256 = "sha256-e79IpLAMCHllSgg9w8tWn4d/Zv+WVDPPjOEUF3+3LFc=";
  ldflags = [ "-s" "-w" ];
}
