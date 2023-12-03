{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorHash = "sha256-lG5KhRjLcsC9Jg+z04mvd668OxmyCvfalhi+o9DvveQ=";
  ldflags = [ "-s" "-w" ];
}
