# Generate protobuf code
FROM znly/protoc AS protoc
WORKDIR /go/src/github.com/ispras/michman/
COPY . .
RUN cd internal/protobuf && /bin/sh generate.sh

# Build launcher
FROM golang:1.16 AS builder
WORKDIR /go/src/github.com/ispras/michman/
COPY --from=protoc /go/src/github.com/ispras/michman/ .
RUN go build -o launch ./cmd/launcher

# Setup runtime
FROM ubuntu:18.04 AS runtime
WORKDIR /go/src/github.com/ispras/michman/
COPY --from=builder /go/src/github.com/ispras/michman/ .

# Install software-properties-common
RUN apt-get update && apt-get install -y \
  software-properties-common

# Install python3.7 and pip
RUN add-apt-repository ppa:deadsnakes/ppa && \
  apt-get update && apt-get install -y \
    python3.7 \
    python3-pip

# Install ssh and rsync
RUN apt-get install -y \
  openssh-server \
  rsync \
  && rm -rf /var/lib/apt/lists/*

# Install pip packages
RUN pip3 install       \
  ansible==2.9.4       \
  openstacksdk==0.40.0 \
  decorator==4.4.2

# Set environment variables
ENV CONFIG ./configs/config.yaml
ENV PORT 5000

# Start launch service
CMD ./launch --config ${CONFIG} --port ${PORT}

# Expose launcher port
EXPOSE ${PORT}

