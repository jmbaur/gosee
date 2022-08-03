{
  description = "gosee";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = inputs: with inputs; {
    overlays.default = final: prev: {
      gosee = prev.buildGoModule {
        pname = "gosee";
        version = "0.1.0";
        src = ./.;
        CGO_ENABLED = 0;
        vendorSha256 = "sha256-0pmE22lo4mxYuAluCnXTliNVLacGxAMLmmr7W5ex+uI=";
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
        buildInputs = with pkgs; [ go_1_18 entr ];
        CGO_ENABLED = 0;
      };
      packages.gosee = pkgs.gosee;
      packages.default = pkgs.gosee;
      apps.gosee = flake-utils.lib.mkApp { drv = pkgs.gosee; name = "gosee"; };
      apps.default = apps.gosee;
    });

}
