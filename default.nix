{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorHash = "sha256-rFmdJpaz2u7dY0DagpQLyXg2V+/PCb2HzbeTEcJFmZI=";
  ldflags = [ "-s" "-w" ];
}
