{
  description = "gosee";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = inputs: with inputs; {
    overlays.default = _: prev: { gosee = prev.callPackage ./. { }; };
  } // flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ self.overlays.default ];
      };
    in
    rec {
      devShells.default = pkgs.mkShell {
        inherit (pkgs.gosee) CGO_ENABLED;
        buildInputs = with pkgs; [ go ];
      };
      packages.default = pkgs.gosee;
      apps.default = { type = "app"; program = "${pkgs.gosee}/bin/gosee"; };
    });

}
