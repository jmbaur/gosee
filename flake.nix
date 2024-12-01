{
  description = "gosee";
  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";
  outputs =
    inputs:
    let
      forAllSystems =
        cb:
        inputs.nixpkgs.lib.genAttrs [ "aarch64-linux" "x86_64-linux" "aarch64-darwin" "x86_64-darwin" ] (
          system:
          cb {
            inherit system;
            pkgs = import inputs.nixpkgs {
              inherit system;
              overlays = [ inputs.self.overlays.default ];
            };
          }
        );
    in
    {
      overlays.default = final: prev: {
        gosee = prev.callPackage ./package.nix { };
        vimPlugins = prev.vimPlugins // {
          gosee-nvim = prev.vimUtils.buildVimPlugin {
            pname = "gosee-nvim";
            version = final.gosee.version;
            src = ./nvim;
          };
        };
      };
      devShells = forAllSystems (
        { pkgs, ... }:
        {
          default = pkgs.mkShell {
            inputsFrom = [ pkgs.gosee ];
            nativeBuildInputs = [ pkgs.just ];
          };
        }
      );
      packages = forAllSystems (
        { pkgs, ... }:
        {
          default = pkgs.gosee;
        }
      );
      apps = forAllSystems (
        { pkgs, ... }:
        {
          default = {
            type = "app";
            program = "${pkgs.gosee}/bin/gosee";
          };
        }
      );
    };
}
