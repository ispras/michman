FROM ubuntu:18.04 as base

#install software-properties-common
RUN apt-get update && \
    apt-get install -y software-properties-common && \
    rm -rf /var/lib/apt/lists/*

#install go
RUN apt-get update && \
	add-apt-repository ppa:longsleep/golang-backports && \
	apt-get -y install golang-go && \
	apt-get -y install unzip apt-transport-https && \
  	apt-get -y install ca-certificates curl software-properties-common

#configure Go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin

# install protobuf from source
RUN apt-get update && \
    apt-get -y install git unzip build-essential autoconf libtool
ADD https://github.com/protocolbuffers/protobuf/releases/download/v3.12.3/protobuf-all-3.12.3.zip .
RUN unzip protobuf-all-3.12.3.zip && \
    cd protobuf-3.12.3 && \
    ./configure && \
    make && \
    make check && \
    make install && \
    ldconfig && \
    make clean

# Get the source from GitHub
RUN go get google.golang.org/grpc

# Get go packages
RUN go get github.com/golang/protobuf/protoc-gen-go && \
  go get google.golang.org/protobuf/reflect/protoreflect && \
  go get google.golang.org/protobuf/runtime/protoimpl && \
	go get github.com/golang/mock/mockgen

#create michman directory
RUN mkdir -p ${GOPATH}/src/github.com/ispras && \
	mkdir -p ${GOPATH}/src/github.com/ispras/michman && \
    mkdir -p /logs

#set workdir
WORKDIR ${GOPATH}/src/github.com/ispras/michman
COPY . .

#generate proto files and compile michman
RUN ["/bin/bash","./build.sh","proto"]
RUN ["/bin/bash","./build.sh","compile"]

#set environment variables
ENV CONFIG ./configs/config.yaml
ENV LAUNCHER localhost:5000
ENV PORT 8081

#start http-server
CMD ./http --config ${CONFIG} --launcher ${LAUNCHER} --port ${PORT}

#expose rest port
EXPOSE ${PORT}
