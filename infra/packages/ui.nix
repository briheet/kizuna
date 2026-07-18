{
  buildNpmPackage,
}:

buildNpmPackage {
  pname = "kizuna-ui";
  version = "0.1.0";

  src = ../../ui;

  npmDepsHash = "sha256-Z/34V2cMkzGZ4Vyh1fqwT5B2xH+DQgetM0Ra7VHTGRM=";

  npmBuildScript = "build";

  installPhase = ''
    runHook preInstall
    mkdir -p "$out/share/kizuna-ui"
    cp -R dist/. "$out/share/kizuna-ui/"
    runHook postInstall
  '';
}
