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
  vendorHash = "sha256-SiA1bJDx+qtT8RhLUuPrFXXJRBlyPU0N9I3crz2gFnA=";
  ldflags = [
    "-s"
    "-w"
  ];
}
