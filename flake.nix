{
  description = "gosee";

  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system}; in
      rec {
        devShell = pkgs.mkShell { buildInputs = with pkgs; [ git go entr ]; };
        packages.gosee = pkgs.buildGoModule {
          name = "gosee";
          src = builtins.path { path = ./.; };
          vendorSha256 = "17icimk6gnhxbv0dc5h9ld6v9ksq63i5zhh0yv5j7hh43j84abb8";
        };
        defaultPackage = packages.gosee;
      });
}
