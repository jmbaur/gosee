{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorSha256 = "sha256-UKN11AJUkddISlsbTV2NDVOiVsYzG5Wzj1eD80t7p64=";
  ldflags = [ "-s" "-w" ];
}
