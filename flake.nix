{
  description = "gosee";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }@inputs: {
    overlays.default = final: prev: {
      gosee = nixpkgs.legacyPackages.${prev.system}.buildGo117Module {
        pname = "gosee";
        version = "0.1.0";
        src = builtins.path { path = ./.; };
        CGO_ENABLED = 0;
        vendorSha256 = "sha256-ISlKGEdypPxKUB7eht4Wj+zLdTA1z1tPvBE4vsVaEyU=";
      };
    };
  } //
  flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        overlays = [ self.overlays.default ];
        inherit system;
      };
    in
    rec {
      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [ git go_1_18 entr ];
        CGO_ENABLED = 0;
      };
      packages.gosee = pkgs.gosee;
      packages.default = pkgs.gosee;
      apps.gosee = flake-utils.lib.mkApp { drv = pkgs.gosee; name = "gosee"; };
      apps.default = apps.gosee;
    });

}
