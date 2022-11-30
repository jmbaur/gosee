{
  description = "gosee";
  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";
  outputs = inputs: with inputs;
    let
      forAllSystems = cb: nixpkgs.lib.genAttrs [ "aarch64-linux" "x86_64-linux" "aarch64-darwin" "x86_64-darwin" ] (system: cb {
        inherit system;
        pkgs = import nixpkgs { inherit system; overlays = [ self.overlays.default ]; };
      });
    in
    {
      overlays.default = final: prev: {
        gosee = prev.callPackage ./. { };
        vimPlugins = prev.vimPlugins // {
          gosee-nvim = prev.vimUtils.buildVimPlugin {
            pname = "gosee-nvim";
            version = final.gosee.version;
            src = ./nvim;
          };
        };
      };
      devShells = forAllSystems ({ pkgs, ... }: {
        default = pkgs.mkShell {
          buildInputs = with pkgs; [ nix-prefetch just ];
          inherit (pkgs.gosee) nativeBuildInputs;
        };
      });
      packages = forAllSystems ({ pkgs, ... }: { default = pkgs.gosee; });
      apps = forAllSystems ({ pkgs, ... }: { default = { type = "app"; program = "${pkgs.gosee}/bin/gosee"; }; });
    };
}
