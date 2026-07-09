{
  buildGoModule,
}:

buildGoModule {
  pname = "kizuna-backend";
  version = "0.1.0";
  src = ../../services/backend;

  vendorHash = "sha256-5DaCYCQohLRp8fg+qJwA461TQE66h7L+wlyAks6VrWM=";

  subPackages = [ "cmd/backend" ];

  ldflags = [
    "-s"
    "-w"
  ];

}
