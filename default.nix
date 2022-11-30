{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorSha256 = "sha256-d+2496rh4ezz0HMZPoGC1z94VavUrT5+4eDufqeOX3A=";
  ldflags = [ "-s" "-w" ];
}
