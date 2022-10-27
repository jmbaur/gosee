help:
	@just --list

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
