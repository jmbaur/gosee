{ buildGoModule, writeShellScriptBin }:
let
  gosee = buildGoModule {
    pname = "gosee";
    version = "0.1.0";
    src = ./.;
    CGO_ENABLED = 0;
    vendorSha256 = "sha256-Zd0YRadV8Gfy2dzP2b9nqZQsR4rXedu+1IHEoYDuzmQ=";
    passthru.update = writeShellScriptBin "update" ''
      if [[ $(${gosee.go}/bin/go get -u ./...) != "" ]]; then
        sed -i 's/vendorSha256\ =\ "sha256-.*";/vendorSha256="sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";/' default.nix
        echo "run 'nix build' then update the vendorSha256 field with the correct value"
      fi
    '';
  };
in
gosee
