{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorHash = "sha256-+47rPmgWVmL3+xYbZsaCGmBB7X5LD3k6gBz5/6ltxNU=";
  ldflags = [ "-s" "-w" ];
}
