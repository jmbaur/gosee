{ pkgs ? import <nixpkgs> {} }:
pkgs.buildGoModule {
  name = "gosee";
  src = builtins.path { path = ./.; };
  vendorSha256 = "17icimk6gnhxbv0dc5h9ld6v9ksq63i5zhh0yv5j7hh43j84abb8";
}
