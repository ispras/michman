# Generate protobuf code
FROM znly/protoc AS protoc
WORKDIR /go/src/github.com/ispras/michman/
COPY . .
RUN /bin/sh ./build.sh proto

# Build launcher
FROM golang:1.16 AS builder
WORKDIR /go/src/github.com/ispras/michman/
COPY --from=protoc /go/src/github.com/ispras/michman/ .
RUN go build -o launch ./cmd/launcher

# Setup runtime
FROM centos:centos8 AS runtime
WORKDIR /go/src/github.com/ispras/michman/
COPY --from=builder /go/src/github.com/ispras/michman/ .

# Install crypto-policies
RUN dnf update -y && \
    dnf install crypto-policies -y

# Openstack Liberty fix
RUN update-crypto-policies --set LEGACY

# Install python3.7 and pip
RUN dnf update -y && \
    dnf install python3 -y

# Install pip packages
RUN pip3 install       \
  setuptools-rust      \
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

