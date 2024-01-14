{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorHash = "sha256-v1PWx4mN/y5Tm9ZpHZwjFcpgWclRl3UmUh6qveIRoqY=";
  ldflags = [ "-s" "-w" ];
}
