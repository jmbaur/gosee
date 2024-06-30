{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorHash = "sha256-sotvEGmBhB//N8JfaQ5N+3fyBDAjJJediyQgHDczbNk=";
  ldflags = [ "-s" "-w" ];
}
