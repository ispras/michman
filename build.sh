#!/bin/bash

usage="usage: build.sh [OPTION...] [COMMAND] \n
Available commands:\n
 proto - creates required protobuf code\n
 mock - creates required mocks for testing\n
 test - run test\n
 compile - build binary files\n
 start - run michman (launcher and rest api)\n
 \tAvailable options:\n
 \t     -c|--config - config path\n
 \t     -l|--launcher-port - launcher port (default vaule is 5000)\n
 \t     -r|--rest-port|--http-port - http port (default value is 8081)\n
 help - show this message"

LAUNCHER_BIN=launch
REST_BIN=http
CONFIG=./config.yaml
LAUNCHER_PORT=5000
HTTP_PORT=8081
while [[ $# -gt 0 ]]
do
key=$1
case $key in
	-c|--config)
		CONFIG=$2
		shift
		shift
		;;
	-l|--launcher-port)
		LAUNCHER_PORT=$2
		shift
		shift
		;;
	-r|--rest-port|--http-port)
		HTTP_PORT=$2
		shift
		shift
		;;
	-h|--help)
		COMMAND=help
		break
		;;
	proto|mock|test|compile|start|stop|clean|help)
		if [[ -n $COMMAND ]]
		then
			echo $usage
			break
		fi
		COMMAND=$key
		shift
		;;
	*)
		echo $usage
		break
		;;
esac
done

case $COMMAND in
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
	1>/dev/null 2>/dev/null ./$LAUNCHER_BIN --config $CONFIG --port $LAUNCHER_PORT &
	echo $! > .$LAUNCHER_BIN.pid
	sleep 1
	if [ -z $(ps | grep $(cat .$LAUNCHER_BIN.pid)| awk '{print $1}') ]
	then 
		echo "launcher did't start, check config and logs"
		rm .$LAUNCHER_BIN.pid
		exit
	fi
	echo "run rest api server..."
	1>/dev/null 2>/dev/null ./$REST_BIN --config $CONFIG --port $HTTP_PORT --launcher localhost:$LAUNCHER_PORT &
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
help) echo -e $usage
#*) echo $usage
esac

