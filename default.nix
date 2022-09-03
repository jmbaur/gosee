{ buildGoModule }:
buildGoModule {
  pname = "gosee";
  version = "0.1.0";
  src = ./.;
  CGO_ENABLED = 0;
  vendorSha256 = "sha256-0pmE22lo4mxYuAluCnXTliNVLacGxAMLmmr7W5ex+uI=";
}
