{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorSha256 = "sha256-EUlqmPOl8rshRdeQQElba/XCX21W9m2wSvYZFFMwiUk=";
  ldflags = [ "-s" "-w" ];
}
