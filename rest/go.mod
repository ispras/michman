module github.com/ispras/michman/rest

replace github.com/ispras/michman/database => ../database

replace github.com/ispras/michman/utils => ../utils

replace github.com/ispras/michman/protobuf => ../protobuf

replace github.com/ispras/michman/mocks => ../mocks

go 1.14

require (
	github.com/golang/mock v1.4.3
	github.com/google/uuid v1.1.1
	github.com/ispras/michman/database v0.0.0-00010101000000-000000000000
	github.com/ispras/michman/mocks v0.0.0-00010101000000-000000000000
	github.com/ispras/michman/protobuf v0.0.0-00010101000000-000000000000
	github.com/ispras/michman/utils v0.0.0-00010101000000-000000000000
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/julienschmidt/httprouter v1.3.0
	google.golang.org/grpc v1.29.1
)
