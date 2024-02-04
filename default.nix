{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorHash = "sha256-QeKvtH1aEO6QfmRmBFO952S0jl04HlLtWM0MtGxZNI4=";
  ldflags = [ "-s" "-w" ];
}
