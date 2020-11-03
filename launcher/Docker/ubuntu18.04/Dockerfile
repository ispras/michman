FROM ubuntu:18.04 as base

#install software-properties-common
RUN apt-get update && \
    apt-get install -y software-properties-common && \
    rm -rf /var/lib/apt/lists/*

#install python3.7 and pip
RUN apt-get update && \
	add-apt-repository ppa:deadsnakes/ppa && \
	apt-get -y install python3.7 && \
	apt-get -y install python3-pip

#install pip packages
RUN pip3 install ansible==2.9.4 openstacksdk==0.40.0

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
	go get github.com/golang/mock/mockgen

#clone michman from github
RUN git clone https://github.com/ispras/michman.git

#move michman
RUN mkdir -p ${GOPATH}/src/github.com/ispras && \
	mv ./michman  ${GOPATH}/src/github.com/ispras/

#create gitlab directory
RUN mkdir -p ${GOPATH}/src/gitlab.at.ispras.ru && \
	mkdir -p ${GOPATH}/src/gitlab.at.ispras.ru/michman

#set workdir
WORKDIR ${GOPATH}/src/gitlab.at.ispras.ru/michman
COPY ../.. .

#generate proto files and compile michman
RUN ["/bin/sh","./build.sh","proto"]
RUN ["/bin/sh","./build.sh","compile"]

#set environment variables
ENV CONFIG /config.yaml
ENV PORT 5000

#start launch service
CMD ./launch --config ${CONFIG} --port ${PORT}

#expose launcher port
EXPOSE ${PORT}

