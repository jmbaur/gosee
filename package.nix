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
  vendorHash = "sha256-2JkuOdnbfaKpP6ZcZYQSqhkS0dDXWJ0hFGPpw/oo8rE=";
  ldflags = [
    "-s"
    "-w"
  ];
}
