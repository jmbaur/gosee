{ buildGoModule, ... }:
buildGoModule {
  pname = "gosee";
  version = "0.2.2";
  src = ./.;
  vendorHash = "sha256-8yk9prK/zf1jkSMQTt7QrEGbyG6Y4NFLjI0+ph25zPY=";
  ldflags = [ "-s" "-w" ];
}
