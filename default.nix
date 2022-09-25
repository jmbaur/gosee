{ buildGoModule
, github-markdown-css
, CGO_ENABLED
, ...
}:
let
  drv = buildGoModule {
    pname = "gosee";
    version = "0.2.0";
    src = ./.;
    vendorSha256 = "sha256-guuJ5NZk7+Ya+oRcvg2paDIagEpKSrcqYJwoorCmb4U=";
    # this cannot be a symlink since go:embed will not read symlinks
    preBuild = ''
      if [[ ! -f static/github-markdown.css ]]; then
        cp ${github-markdown-css}/github-markdown.css static/github-markdown.css
      fi
    '';
    inherit CGO_ENABLED;
  };
in
drv
