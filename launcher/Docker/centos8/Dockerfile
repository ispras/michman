FROM centos:centos8 as base

# Openstack Liberty fix
RUN update-crypto-policies --set LEGACY

#install go
ADD https://dl.google.com/go/go1.13.5.linux-amd64.tar.gz .
RUN sha256sum go1.13.5.linux-amd64.tar.gz && \
	tar -C /usr/local -xf go1.13.5.linux-amd64.tar.gz

#configure Go
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH /usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin:/root/bin
ENV PATH ${GOPATH}/bin:/usr/local/go/bin:$PATH


RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin

#install development tools
RUN yum -y update && \
    yum -y install git unzip autoconf libtool && \
    yum -y groupinstall "Development Tools"

#install python3.7 and pip
RUN dnf update -y && \
    dnf install python3 -y

#install pip packages
RUN pip3 install ansible==2.9.4 openstacksdk==0.40.0

# install protobuf
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

#create michman directory
RUN mkdir -p ${GOPATH}/src/github.com/ispras && \
	mkdir -p ${GOPATH}/src/github.com/ispras/michman

#set workdir
WORKDIR ${GOPATH}/src/github.com/ispras/michman
COPY . .


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
