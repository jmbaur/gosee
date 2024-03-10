{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorHash = "sha256-Rq7w0qtllbdDwidVJCGLBUbLTtQ/0KYjZLV2KApez4M=";
  ldflags = [ "-s" "-w" ];
}
