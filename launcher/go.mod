module github.com/ispras/michman/launcher

replace github.com/ispras/michman/database => ../database

replace github.com/ispras/michman/utils => ../utils

replace github.com/ispras/michman/protobuf => ../protobuf

replace github.com/ispras/michman/logger => ../logger

go 1.14

require (
	github.com/hashicorp/vault/api v1.0.4
	github.com/ispras/michman/database v0.0.0-00010101000000-000000000000
	github.com/ispras/michman/logger v0.0.0-00010101000000-000000000000
	github.com/ispras/michman/protobuf v0.0.0-00010101000000-000000000000
	github.com/ispras/michman/utils v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.29.1
)
