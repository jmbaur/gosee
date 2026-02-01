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
  vendorHash = "sha256-lJ4TL4NnNKrA2Ocj93zJI19EjsKKzX1eMnefto5gwn4=";
  ldflags = [
    "-s"
    "-w"
  ];
}
