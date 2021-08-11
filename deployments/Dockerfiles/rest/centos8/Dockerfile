FROM centos:centos8 as base

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

# install protobuf from source
#RUN yum -y install epel-release && \
#	yum -y install snapd && \
#	systemctl enable --now snapd.socket && \
#	ln -y -s /var/lib/snapd/snap /snap && \
#	snap -y install protobuf --classic

RUN yum -y update && \
    yum -y install git unzip autoconf libtool && \
    yum -y groupinstall "Development Tools"

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
	mkdir -p ${GOPATH}/src/github.com/ispras/michman

#set workdir
WORKDIR ${GOPATH}/src/github.com/ispras/michman
COPY . .

#generate proto files and compile michman
RUN ["/bin/sh","./build.sh","proto"]
RUN ["/bin/sh","./build.sh","compile"]

#set environment variables
ENV CONFIG ./configs/config.yaml
ENV LAUNCHER localhost:5000
ENV PORT 8081

#start http-server
CMD ./http --config ${CONFIG} --launcher ${LAUNCHER} --port ${PORT}

#expose rest port
EXPOSE ${PORT}
