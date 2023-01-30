#!/bin/bash

usage="usage: build.sh [OPTION...] [COMMAND] \n
Available commands:\n
 proto - creates required protobuf code. Available options:\n
 \t     -d|--docker - set USE_DOCKER=true
 mock - creates required mocks for testing\n
 test - run test\n
 compile - build binary files\n
 clean - remove all generated files\n
 reset - stop processes and remove all generated files\n
 start - run michman (launcher and rest api). Available options:\n
 \t     -c|--config - config path\n
 \t     -l|--launcher-port - launcher port (default value is 5000)\n
 \t     -r|--rest-port|--http-port - http port (default value is 8081)\n
 status - check launch and http processes\n
 help - show this message"

LAUNCHER_BIN=launch
REST_BIN=http
LAUNCHER_START_LOG=.launch_start.log
REST_START_LOG=.http_start.log
CONFIG=./configs/config.yaml
LAUNCHER_PORT=5000
HTTP_PORT=8081
PROTO_CODE=./internal/protobuf/launcher.pb.go

MOCK_DATABASE=mock_database.go
MOCK_GRPC=mock_grpcclient.go
MOCK_VAULT=mock_vault.go
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
  -d|--docker)
		USE_DOCKER=true
		shift
		shift
		;;
	-h|--help)
		COMMAND=help
		break
		;;
	proto|mock|test|compile|start|restart|reset|status|stop|clean|help)
		if [[ -n "$COMMAND" ]]
		then
			echo $usage
			break
		fi
		COMMAND=$key
		shift
		;;
	*)
		echo -e $usage
		break
		;;
esac
done

function generate_proto() {
  if [ "$USE_DOCKER" == true ]
  then
    if ! command -v docker &> /dev/null
    then
        echo "ERROR: docker is not installed or not running"
        exit 1
    else
      echo "generate protobuf code... (docker)"
      cd internal/protobuf
      docker run --platform linux/amd64\
        --rm \
        -v $(pwd)/../../:$(pwd)/../../ \
        -w $(pwd) znly/protoc \
        --go_out=paths=source_relative,plugins=grpc:. \
        -I. \
        launcher.proto || result=$?
      cd ../..
      if [ "$result" == 1 ]
      then
        exit 1
      fi
    fi
  else
    if ! command -v protoc &> /dev/null
    then
        echo "ERROR: protoc is not installed. Use -d (--docker) flag to use protoc in docker"
        exit 1
    else
      echo "generate protobuf code..."
      cd internal/protobuf
      protoc \
        --go_out=paths=source_relative,plugins=grpc:. \
        --go_opt=Mlauncher.proto=github.com/ispras/michman/internal/protobuf \
        launcher.proto
      cd ../..
    fi
  fi
}

function generate_mock() {
  if [ -z $( 2>/dev/null ls $PROTO_CODE ) ]
  then
    generate_proto
  fi

  echo "generate mocks..."
  if ! command -v mockgen &> /dev/null
  then
      echo "ERROR: mockgen is not installed"
      exit 1
  else
    cd ./internal/database
    mockgen --destination=../mock/$MOCK_DATABASE -package=mock . Database
    cd ../..

    cd ./internal/rest/handler
    mockgen --destination=../../mock/$MOCK_GRPC -package=mock . GrpcClient
    cd ../../..

    cd ./internal/utils
    mockgen --destination=../mock/$MOCK_VAULT -package=mock . SecretStorage
    cd ../..
  fi
}

function run_tests() {
  if [ -z $( 2>/dev/null ls ./internal/mocks/$MOCK_DATABASE ./internal/mocks/$MOCK_GRPC ./internal/mocks/$MOCK_VAULT ) ]
  then
    generate_mock
  fi

  #TODO: all tests
  echo "run tests..."
  cd ./test/handlers
  go test
}

function compile() {
  if [ -z $( 2>/dev/null ls $PROTO_CODE ) ]
  then
   generate_proto
  fi
  echo "build launcher..."
  go build -o $LAUNCHER_BIN ./cmd/launcher
  echo "build rest api server..."
  go build -o $REST_BIN ./cmd/rest
}

function start() {
  if [[ -f ".$LAUNCHER_BIN.pid" ]] || [[ -f ".$REST_BIN.pid" ]]
  then
    echo "Michman is already running"
    exit 0
  fi

  if [ -z $( 2>/dev/null ls $PROTO_CODE ) ]
  then
    generate_proto
  fi

  if [ -z $(test -f "$LAUNCHER_BIN") ] || [ -z $(test -f "$REST_BIN") ]
  then
    compile
  fi

  echo "run launcher..."
  ./$LAUNCHER_BIN --config $CONFIG --port $LAUNCHER_PORT 1>/dev/null 2>$LAUNCHER_START_LOG &
  echo $! > .$LAUNCHER_BIN.pid
  sleep 1

  if [[ -z $(ps -eo pid | grep -w $(cat .$LAUNCHER_BIN.pid)) ]]
  then
    echo "ERROR: launcher didn't start, check launch_start.log file for more information"
    rm .$LAUNCHER_BIN.pid
    exit
  fi

  echo "run rest api server..."
  ./$REST_BIN --config $CONFIG --port $HTTP_PORT --launcher localhost:$LAUNCHER_PORT 1>/dev/null 2>$REST_START_LOG &
  echo $! > .$REST_BIN.pid
  sleep 1

  if [[ -z $(ps -eo pid | grep -w $(cat .$REST_BIN.pid)) ]]
  then
    echo "ERROR: rest api didn't start, check http_start.log file for more information"
    rm .$REST_BIN.pid
    exit
  fi
}

function restart() {
    stop
    start
}

function stop() {
  if [[ -f ".$LAUNCHER_BIN.pid" ]]
  then
    echo "kill launcher..."
    if ps -p $( cat .$LAUNCHER_BIN.pid ) > /dev/null
    then
      kill $( cat .$LAUNCHER_BIN.pid )
    fi
    rm .$LAUNCHER_BIN.pid
  fi

  if [[ -f ".$REST_BIN.pid" ]]
  then
    echo "kill rest api server..."
    if ps -p $( cat .$REST_BIN.pid ) > /dev/null
    then
      kill $( cat .$REST_BIN.pid )
    fi
    rm .$REST_BIN.pid
  fi
}

function clean() {
  echo "remove all generated and binary files:"

  if test -f ./$LAUNCHER_BIN; then
    rm -rf ./$LAUNCHER_BIN
    echo "rm launch"
  fi

  if test -f ./$REST_BIN; then
    rm -rf ./$REST_BIN
    echo "rm http"
  fi

  if test -f $PROTO_CODE; then
    rm -rf $PROTO_CODE
    echo "rm proto files"
  fi

  if test -f ./internal/mock/$MOCK_DATABASE; then
    rm -rf ./internal/mock/mock_*
    echo "rm mock files"
  fi

  if test -f ./$LAUNCHER_START_LOG; then
    rm -rf ./$LAUNCHER_START_LOG
    echo "rm launch_start.log"
  fi

  if test -f ./$REST_START_LOG; then
    rm -rf ./$REST_START_LOG
    echo "rm http_start.log"
  fi
}

function reset() {
    stop
    clean
}

function status() {
  launch_running=0
  launch_running=0
  if [[ -f ".$LAUNCHER_BIN.pid" ]]
  then
      if ps -p $( cat .$LAUNCHER_BIN.pid ) > /dev/null
      then
        echo "launch process is alive (pid: $( cat .$LAUNCHER_BIN.pid ))"
      else
        echo "launch process with pid $( cat .$LAUNCHER_BIN.pid ) does not exist any more"
      fi
  else
    launch_running=1
  fi

  if [[ -f ".$REST_BIN.pid" ]]
  then
    if ps -p $( cat .$REST_BIN.pid ) > /dev/null
    then
      echo "http process is alive (pid: $( cat .$REST_BIN.pid ))"
    else
      echo "http process with pid $( cat .$REST_BIN.pid ) does not exist any more"
    fi
  else
    http_running=1
  fi

  if [[ $(( launch_running + http_running )) == 2 ]]
  then
    echo "michman is not running"
  fi
}

case $COMMAND in
proto)
	generate_proto
	;;
mock)
  generate_mock
  ;;
test)
  run_tests
	;;
compile)
  compile
	;;
start)
  start
	;;
restart)
  restart
  ;;
status)
	status
	;;
stop)
	stop
	;;
clean)
	clean
	;;
reset)
	reset
	;;
help)
  echo -e $usage
  ;;
esac