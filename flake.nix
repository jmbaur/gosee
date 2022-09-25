{
  description = "gosee";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    github-markdown-css.flake = false;
    github-markdown-css.url = "github:sindresorhus/github-markdown-css";
    nixpkgs.url = "nixpkgs/nixos-unstable";
  };

  outputs = inputs: with inputs; {
    overlays.default = _: prev: {
      gosee = prev.callPackage ./. {
        inherit github-markdown-css;
        CGO_ENABLED = 0;
      };
    };
  } // flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ self.overlays.default ];
      };
    in
    rec {
      devShells.default = pkgs.mkShell {
        inherit (pkgs.gosee) nativeBuildInputs CGO_ENABLED;
        shellHook = pkgs.gosee.preBuild;
      };
      packages.default = pkgs.gosee;
      apps.default = {
        type = "app";
        program = "${pkgs.gosee}/bin/gosee";
      };
    });

}
