{
  description = "gosee";

  inputs.nixpkgs.url = "nixpkgs/nixos-21.11";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }@inputs:
    flake-utils.lib.eachDefaultSystem
      (system:
        let pkgs = nixpkgs.legacyPackages.${system}; in
        rec {
          devShell = pkgs.mkShell { buildInputs = with pkgs; [ git go entr ]; };
          packages.gosee = pkgs.buildGoModule {
            name = "gosee";
            src = builtins.path { path = ./.; };
            CGO_ENABLED = 0;
            vendorSha256 = "17icimk6gnhxbv0dc5h9ld6v9ksq63i5zhh0yv5j7hh43j84abb8";
          };
          defaultPackage = packages.gosee;
        })
    //
    {
      overlay = self: super: { gosee = self.packages.gosee; };
    };
}
