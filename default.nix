{ buildGoModule, writeShellScriptBin }:
let
  drv = buildGoModule {
    pname = "gosee";
    version = "0.1.0";
    src = ./.;
    CGO_ENABLED = 0;
    vendorSha256 = "sha256-Zd0YRadV8Gfy2dzP2b9nqZQsR4rXedu+1IHEoYDuzmQ=";
    passthru.update = writeShellScriptBin "update" ''
      if [[ $(${drv.go}/bin/go get -u all 2>&1) != "" ]]; then
        sed -i 's/vendorSha256\ =.*;/vendorSha256="sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";/' default.nix
        ${drv.go}/bin/go mod tidy
      fi
    '';
  };
in
drv
