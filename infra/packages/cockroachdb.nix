{
  lib,
  stdenv,
  fetchzip,
  buildFHSEnv,
}:

let
  version = "25.4.12";
  srcs = {
    x86_64-linux = fetchzip {
      url = "https://binaries.cockroachdb.com/cockroach-v${version}.linux-amd64.tgz";
      hash = "sha256-Zjm6rxfk8fpjOFmmvddOvBklhi3zV4jnLtK6MGuxHK4=";
    };
  };
  src =
    srcs.${stdenv.hostPlatform.system}
      or (throw "Unsupported CockroachDB platform: ${stdenv.hostPlatform.system}");
in
buildFHSEnv {
  pname = "cockroachdb";
  inherit version;

  runScript = "${src}/cockroach";

  extraInstallCommands = ''
    cp -P "$out/bin/cockroachdb" "$out/bin/cockroach"
  '';

  meta = {
    description = "Scalable, survivable, strongly consistent SQL database";
    homepage = "https://www.cockroachlabs.com";
    license = lib.licenses.unfree;
    sourceProvenance = [ lib.sourceTypes.binaryNativeCode ];
    platforms = [ "x86_64-linux" ];
  };
}
