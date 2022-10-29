help:
	@just --list

update:
	#!/usr/bin/env bash
	go get -u all
	export NIX_PATH="nixpkgs=$(nix flake prefetch nixpkgs --json | jq --raw-output '.storePath')"
	newvendorSha256="$(nix-prefetch \
		 "{ sha256 }: ((import <nixpkgs> {}).callPackage ./. {}).go-modules.overrideAttrs (_: { vendorSha256 = sha256; })")"
	sed -i "s|vendorSha256.*|vendorSha256 = \"$newvendorSha256\";|" default.nix

build-ui out="static":
	#!/bin/sh
	esbuild index.js \
		--bundle \
		--minify \
		--sourcemap \
		--target=chrome58,firefox57,safari11,edge18 \
		--outdir={{out}}
	cp favicon.webp {{out}}/

build: build-ui
	go build -o $out/gosee .

run: build-ui
	go run .
