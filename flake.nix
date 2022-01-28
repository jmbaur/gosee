{
  description = "gosee";

  inputs.nixpkgs.url = "nixpkgs/nixos-21.11";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }@inputs: {
    overlay = final: prev: {
      gosee = nixpkgs.legacyPackages.${prev.system}.buildGo117Module {
        pname = "gosee";
        version = "0.1.0";
        src = builtins.path { path = ./.; };
        CGO_ENABLED = 0;
        vendorSha256 = "17icimk6gnhxbv0dc5h9ld6v9ksq63i5zhh0yv5j7hh43j84abb8";
      };
    };
  } //
  flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        overlays = [ self.overlay ];
        inherit system;
      };
    in
    rec {
      devShell = pkgs.mkShell { buildInputs = with pkgs; [ git go_1_17 entr ]; };
      packages.gosee = pkgs.gosee;
      defaultPackage = pkgs.gosee;
      apps.gosee = flake-utils.lib.mkApp { drv = pkgs.gosee; name = "gosee"; };
      defaultApp = apps.gosee;
    });


}
