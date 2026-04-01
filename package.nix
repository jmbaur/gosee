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
  vendorHash = "sha256-QB4AbC+JVYfS+Fb2HxmZfXh4lm4pa4/0/5DLgLirSl0=";
  ldflags = [
    "-s"
    "-w"
  ];
}
