{
  buildGoModule,
}:

buildGoModule {
  pname = "kizuna-workers";
  version = "0.1.0";
  src = ../../services/workers;

  vendorHash = "sha256-ssDSGLmCYAiLUbIo8j9dMR6pZ+EWAG4DkBypDtTNUtw=";

  subPackages = [ "cmd/workers" ];

  ldflags = [
    "-s"
    "-w"
  ];
}
