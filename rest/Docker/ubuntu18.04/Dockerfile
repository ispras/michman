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

# Install protoc-gen-go
RUN go get github.com/golang/protobuf/protoc-gen-go

#install Go packages
RUN go get -u github.com/hashicorp/vault/api  && \
	go get -u gopkg.in/yaml.v2  && \
	go get github.com/julienschmidt/httprouter  && \
	go get -u gopkg.in/couchbase/gocb.v1  && \
	go get github.com/google/uuid  && \
	go get github.com/golang/mock/gomock  && \
	go get github.com/jinzhu/copier && \
	go get github.com/golang/mock/mockgen && \
    go get github.com/alexedwards/scs && \
    go get github.com/casbin/casbin


#create michman directory
RUN mkdir -p ${GOPATH}/src/github.com/ispras && \
	mkdir -p ${GOPATH}/src/github.com/ispras/michman

#set workdir
WORKDIR ${GOPATH}/src/github.com/ispras/michman
COPY . .

#generate proto files and compile michman
RUN ["/bin/bash","./build.sh","proto"]
RUN ["/bin/bash","./build.sh","compile"]

#set environment variables
ENV CONFIG /config.yaml
ENV LAUNCHER localhost:5000
ENV PORT 8081

#start http-server
CMD ./http --config ${CONFIG} --launcher ${LAUNCHER} --port ${PORT}

#expose rest port
EXPOSE ${PORT}
