if [ "$USE_DOCKER" == true ]
then
  docker run \
    --rm \
    -v $(pwd)/../../:$(pwd)/../../ \
    -w $(pwd) znly/protoc \
    --go_out=paths=source_relative,plugins=grpc:. \
    -I. \
    launcher.proto
else
  protoc \
    --go_out=paths=source_relative,plugins=grpc:. \
    --go_opt=Mlauncher.proto=github.com/ispras/michman/internal/protobuf \
    launcher.proto
fi

