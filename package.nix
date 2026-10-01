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
  vendorHash = "sha256-eGiCpH6ocBhddO59H9dV80uJ17OdIMesCcnkLEk/9Kw=";
  ldflags = [
    "-s"
    "-w"
  ];
}
