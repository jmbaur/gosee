{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorHash = "sha256-TBuf9pZNZSTdfMmYjAalCduaApFQxUvNGsirqMfhCQw=";
  ldflags = [ "-s" "-w" ];
}
