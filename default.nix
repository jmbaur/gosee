{ buildGoModule, CGO_ENABLED ? 0, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorSha256 = "sha256-gi1GUaus3r/3P8KBdWndEmHxAXg6vPXnQysGBezO0rQ=";
  ldflags = [ "-s" "-w" ];
  inherit CGO_ENABLED;
}
