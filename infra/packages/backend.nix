{
  buildGoModule,
}:

buildGoModule {
  pname = "kizuna-backend";
  version = "0.1.0";
  src = ../../services/backend;

  vendorHash = "sha256-Sz3vbsgO2Racw8qUlITrzBZ/Je3vZ+PtT5KZNqejx80=";

  subPackages = [ "cmd/backend" ];

  ldflags = [
    "-s"
    "-w"
  ];

  postInstall = ''
    mkdir -p "$out/share/kizuna-backend/migration"
    cp migration/*.sql "$out/share/kizuna-backend/migration/"
  '';

}
