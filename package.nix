{ lib, buildGoModule }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = lib.fileset.toSource {
    root = ./.;
    fileset = lib.fileset.unions [
      ./go.mod
      ./go.sum
      ./index.html.tmpl
      ./main.go
    ];
  };
  vendorHash = "sha256-+47rPmgWVmL3+xYbZsaCGmBB7X5LD3k6gBz5/6ltxNU=";
  ldflags = [
    "-s"
    "-w"
  ];
}
