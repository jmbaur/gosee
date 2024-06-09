{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorHash = "sha256-8tyhb5WGgNGQ6xk+52v/BAI48uVsoKSMRqKfjKB/zD4=";
  ldflags = [ "-s" "-w" ];
}
