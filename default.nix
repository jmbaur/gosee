{ buildGoModule }:
buildGoModule {
  pname = "gosee";
  version = "0.1.0";
  src = ./.;
  CGO_ENABLED = 0;
  vendorSha256 = "sha256-Zd0YRadV8Gfy2dzP2b9nqZQsR4rXedu+1IHEoYDuzmQ=";
}
