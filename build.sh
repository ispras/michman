#!/bin/sh

usage="usage: build.sh [proto|mock|test|compile|start|stop|clean]"
LAUNCHER_BIN=launch
REST_BIN=http

case $1 in
proto) 
	echo "generate protobuf code..."
	cd protobuf/; protoc --go_out=plugins=grpc:. protofile.proto; cd ..
	;;
mock)
        if [ -z $( 2>/dev/null ls ./protobuf/protofile.pb.go ) ]
        then
                ./build.sh proto
        fi
	echo "generate mocks..."
	cd ./database
	mockgen --destination=../mocks/mock_database.go -package=mocks . Database
        cd ..
        cd ./rest/handlers
        mockgen --destination=../../mocks/mock_grpcclient.go -package=mocks . GrpcClient
        cd ../..
        cd ./utils
        mockgen --destination=../mocks/mock_vault.go -package=mocks . SecretStorage
        cd ..
        ;;
test) 
	if [ -z $( 2>/dev/null ls ./mocks/mock_database.go ./mocks/mock_grpcclient.go ./mocks/mock_vault.go ) ]
	then
		./build.sh mock
	fi
	echo "run tests..."
	cd ./rest/handlers
	go test
	;;
compile)
        if [ -z $( 2>/dev/null ls ./protobuf/protofile.pb.go ) ]
        then
                ./build.sh proto
        fi
	echo "build launcher..."
	cd launcher
	go build
	cd ..
	mv ./launcher/launcher ./$LAUNCHER_BIN
	echo "build rest api server..."
	cd rest
	go build
	cd ..
	mv rest/rest ./$REST_BIN
	;;
start)
	if [ -z $( 2>/dev/null ls ./protobuf/protofile.pb.go ) ]
	then
		./build.sh proto
	fi
	if [ -z $( 2>/dev/null ls ./$LAUNCHER_BIN ) ] || [ -z $( 2>/dev/null ls ./$REST_BIN ) ]
	then
		./build.sh compile
	fi
	echo "run launcher..."
	1>/dev/null 2>/dev/null ./$LAUNCHER_BIN $2 &
	echo $! > .$LAUNCHER_BIN.pid
	sleep 1
	if [ -z $(ps | grep $(cat .$LAUNCHER_BIN.pid)| awk '{print $1}') ]
	then 
		echo "launcher did't start, check config and logs"
		rm .$LAUNCHER_BIN.pid
		exit
	fi
	echo "run rest api server..."
	1>/dev/null 2>/dev/null ./$REST_BIN $2 &
	echo $! > .$REST_BIN.pid
	sleep 1
	if [ -z $(ps | grep $(cat .$REST_BIN.pid) | awk '{print $1}') ]
        then
                echo "rest api did't start, check config and logs"
                rm .$REST_BIN.pid
                exit
        fi
	;;
stop) 
	echo "kill launcher..."
	kill $( cat .$LAUNCHER_BIN.pid )
	rm .$LAUNCHER_BIN.pid
	echo "kill rest api server..."
	kill $( cat .$REST_BIN.pid )
	rm .$REST_BIN.pid
	;;
clean)
	echo "remove all generated and binary files"
	1>/dev/null 2>/dev/null rm ./$LAUNCHER_BIN ./$REST_BIN ./protobuf/protofile.pb.go ./mocks/mock_*
	;;
*) echo $usage
esac

