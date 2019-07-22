# Go server

## ansible-pb
Contains proto file for ansible_service with gRPC and have already generated code from it. Used in ansible_service.

## ansible_service
Contains service for ansible launching. Use ansible-pb for serving requests

## go-server
Server that handles HTTP requests(probably from Envoy, that will take tham from real clients), and call, using gRPC(use client class from ansible-pb), ansible_service to lanch ansible.

# How to get it worked
Launch ansible_service:
```
go run ansible_service/ansible-service.go
```

Launch server:
```
go run go-server.go
```

Send request to localhost:8080/clusters":
```
curl localhost:8080/clusters
```
