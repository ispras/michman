module github.com/ispras/michman/mocks

replace github.com/ispras/michman/protobuf => ../protobuf

replace github.com/ispras/michman/utils => ../utils

go 1.14

require (
	github.com/golang/mock v1.4.3
	github.com/hashicorp/vault/api v1.0.4
	github.com/ispras/michman/protobuf v0.0.0-00010101000000-000000000000
	github.com/ispras/michman/utils v0.0.0-00010101000000-000000000000
)
