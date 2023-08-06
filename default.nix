{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorSha256 = "";
  ldflags = [ "-s" "-w" ];
}
