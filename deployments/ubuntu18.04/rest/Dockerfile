# Generate protobuf code
FROM znly/protoc AS protoc
WORKDIR /go/src/github.com/ispras/michman/
COPY . .
RUN cd internal/protobuf && /bin/sh generate.sh

# Build http server
FROM golang:1.16 AS builder
WORKDIR /go/src/github.com/ispras/michman/
COPY --from=protoc /go/src/github.com/ispras/michman/ ./
RUN go build -o http ./cmd/rest

# Setup runtime
FROM ubuntu:18.04 AS runtime
WORKDIR /go/src/github.com/ispras/michman/
COPY --from=builder /go/src/github.com/ispras/michman/ ./

ARG LAUNCHER_ADDRESS

# Set environment variables
ENV CONFIG ./configs/config.yaml
ENV LAUNCHER $LAUNCHER_ADDRESS
ENV PORT 8081

# Start http-server
CMD ./http --config ${CONFIG} --launcher ${LAUNCHER} --port ${PORT}

# Expose rest port
EXPOSE ${PORT}

