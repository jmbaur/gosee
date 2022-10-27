{ buildGoModule, CGO_ENABLED ? 0, ui-assets, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.1";
  src = ./.;
  vendorSha256 = "sha256-gi1GUaus3r/3P8KBdWndEmHxAXg6vPXnQysGBezO0rQ=";
  preBuild = "cp -r ${ui-assets} static";
  ldflags = [ "-s" "-w" ];
  inherit CGO_ENABLED;
}
