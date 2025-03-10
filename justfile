help:
	@just --list

update:
	#!/usr/bin/env bash
	go get -u all
	go mod tidy
	export NIX_PATH="nixpkgs=$(nix flake prefetch nixpkgs --json | jq --raw-output '.storePath')"
	newvendorHash=$(nix build --impure --expr 'with import <nixpkgs> {}; (callPackage ./package.nix {}).goModules.overrideAttrs (_: {outputHash = ""; outputHashAlgo = "sha256";})' 2>&1 | grep 'got: ' | cut -d':' -f2 | xargs)
	if [[ $newvendorHash != "" ]]; then
		sed -i "s|vendorHash.*|vendorHash = \"$newvendorHash\";|" package.nix
	else
		echo "failed to fetch new vendor hash"
		exit 1
	fi

build:
	go build -o $out/gosee .

run:
	go run .
