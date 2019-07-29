# Http server

## Services

Contains service for ansible launching.

## Protobuf

Contains proto file for gRPC and have already generated code from it. Used in http_server and services/ansible_runner.

## http_server
Server that handles HTTP requests(probably from Envoy, that will take tham from real clients), and call, using gRPC(use client class from ansible-pb), ansible_service to lanch ansible.

# How to get it worked
Launch ansible_runner service:
```
go run src/services/ansible_runner/grpc_server.go
```

Launch http_server:
```
go run src/http_server.go
```

Send request to localhost:8080/clusters":
```
curl localhost:8080/clusters -XPOST -d '{"name": "test", "slaves":3}'
```
