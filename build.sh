#!/bin/sh

usage="usage: build.sh [proto|mock|test|compile|start|stop|clean]"

case $1 in
proto) 
	echo "generate protobuf code..."
	cd src/protobuf/; protoc --go_out=plugins=grpc:. protofile.proto; cd ../..
	;;
mock) 
	echo "generate mocks..."
	mockgen --destination=./src/mocks/mock_database.go -package=mocks github.com/ispras/michman/src/database Database 
	mockgen --destination=./src/mocks/mock_grpcclient.go -package=mocks github.com/ispras/michman/src/handlers GrpcClient 
	mockgen --destination=./src/mocks/mock_vault.go -package=mocks github.com/ispras/michman/src/utils SecretStorage
	;;
test) 
	if [ -z $( 2>/dev/null ls ./src/mocks/mock_database.go ./src/mocks/mock_grpcclient.go ./src/mocks/mock_vault.go ) ]
	then
		./build.sh mock
	fi
	echo "run tests..."
	go test ./src/handlers -cover
	;;
compile) 
	echo "build ansible_service..."
	go build ./src/services/ansible_service/ansible_service.go ./src/services/ansible_service/ansible_launch.go
	echo "build http_server..."
	go build ./src/http_server.go
	;;
start)
	if [ -z $( 2>/dev/null ls ./src/protobuf/protofile.pb.go ) ]
	then
		./build.sh proto
	fi
	if [ -z $( 2>/dev/null ls ./ansible_service ) ] || [ -z $( 2>/dev/null ls ./http_server ) ]
	then
		./build.sh compile
	fi
	echo "run ansible_service..."
	1>/dev/null 2>/dev/null ./ansible_service &
	echo $! > .ansible.pid
	echo "run http_server..."
	1>/dev/null 2>/dev/null ./http_server &
	echo $! > .http.pid
	;;
stop) 
	echo "kill ansible_service..."
	kill $( cat .ansible.pid )
	rm .ansible.pid
	echo "kill http_server..."
	kill $( cat .http.pid )
	rm .http.pid
	;;
clean)
	echo "remove all generated and binary files"
	rm ./ansible_service ./http_server ./src/protobuf/protofile.pb.go ./src/mocks/*
	;;
*) echo $usage
esac

