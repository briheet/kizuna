{
  buildNpmPackage,
}:

buildNpmPackage {
  pname = "kizuna-ui";
  version = "0.1.0";

  src = ../../ui;

  npmDepsHash = "sha256-K+pT8yWFVCjlQk8WCSibUAy/AwX494NkiM1o4JiKtQ4=";

  npmBuildScript = "build";

  installPhase = ''
    runHook preInstall
    mkdir -p "$out/share/kizuna-ui"
    cp -R dist/. "$out/share/kizuna-ui/"
    runHook postInstall
  '';
}
