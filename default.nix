{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorHash = "sha256-Vt0bNJuJ/1djAI4CbaU2rRIvhLPGjzJvkV1j/a9w7pc=";
  ldflags = [ "-s" "-w" ];
}
