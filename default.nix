{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorSha256 = "sha256-/EVloZRF/YGSejuNhl9um4tZiwqZ3tMTkaM84wCVrkQ=";
  ldflags = [ "-s" "-w" ];
}
