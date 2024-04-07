{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorHash = "sha256-+R4t3K/dv8sCuxGYKwkzubL1CTC4mtS9c7bZ+/T9OPs=";
  ldflags = [ "-s" "-w" ];
}
