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
  vendorHash = "sha256-Xy0KEksYdArQ81Od2yXbx0CdA09C3WHBLDGwr2scUzg=";
  ldflags = [
    "-s"
    "-w"
  ];
}
