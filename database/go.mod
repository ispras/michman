module github.com/ispras/michman/database

replace github.com/ispras/michman/protobuf => ../protobuf

replace github.com/ispras/michman/utils => ../utils

go 1.14

require (
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/ispras/michman/protobuf v0.0.0-00010101000000-000000000000
	github.com/ispras/michman/utils v0.0.0-00010101000000-000000000000
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	golang.org/x/net v0.0.0-20200425230154-ff2c4b7c35a0 // indirect
	gopkg.in/couchbase/gocb.v1 v1.6.7
	gopkg.in/couchbase/gocbcore.v7 v7.1.17 // indirect
	gopkg.in/couchbaselabs/gocbconnstr.v1 v1.0.4 // indirect
	gopkg.in/couchbaselabs/jsonx.v1 v1.0.0 // indirect
)
