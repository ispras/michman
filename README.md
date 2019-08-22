# Http server

## Dependencies

gRPC:
```
go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
```

Vault API:
```
go get -u github.com/hashicorp/vault/api
```

YAML:
```
go get -u gopkg.in/yaml.v2
```

httprouting:
```
go get github.com/julienschmidt/httprouter
```

Ansible version >= 2.8.1 

## Configurations
Configuration of vault server is stored in **vault.yaml** file. Example:
```yaml
token: MY-TOKEN
vault_addr: http://127.0.0.1:8200
os_key: kv/openstack
ssh_key: kv/ssh-keys
```

Openstack (os_key) secrets includes following keys:
* **OS_AUTH_URL**
* **OS_PASSWORD**
* **OS_PROJECT_NAME**
* **OS_REGION_NAME**
* **OS_TENANT_ID**
* **OS_TENANT_NAME**
* **OS_USERNAME** 
* **OS_SWIFT_USERNAME** -- optional
* **OS_SWIFT_PASSWORD** -- optional 

Ssh (ssh_key) secrets includes following keys:
* **id_rsa** -- private ssh key for Ansible commands

Configuration of Openstack is stored in **openstack_config.yaml** file. Example:
```yaml
os_key_name: my_key
virtual_network: test
os_image: 731c8c7d-47fd-4b69-bdb4-00415e3ccb00
floating_ip_pool: test
flavor: keystone.medium
```

Where:
* **os_key_name** - key pair name
* **virtual_network** - your virtual network name or ID (in Neutron or Nova-networking)
* **os_image** - ID of OS image
* **floating_ip_pool** - floating IP pool name
* **flavor** - instance flavor that exists in your Openstack environment (e.g. spark.large)
## Services

Contains service for ansible launching.

Supported services types are:
* **cassandra**
* **spark**
* **elastic**
* **jupyter**
* **ignite**
* **jupyterhub** 

Config parameter for **spark** service type supports:
* **use-yarn** -- Spark-on-YARN deploy mode  (has overhead on memory so do not use it if you don't know why)
* **hadoop-version** -- use specific Hadoop version for Spark. Default is the latest supported in Spark.
* **spark-worker-mem-mb** --  don't auto-detect spark worker memory and use specified value, can be useful if other
                             processes on slave nodes (e.g. python) need more memory, default for 10Gb-20Gb RAM slaves is to leave 2Gb to
                             system/other processes; 
* **yarn-master-mem-mb** -- Amount of physical memory, in MB, that can be allocated for containers. Default value if 10240.
                             
Example:
```json
"Config": {
  "use-yarn": "false",
  "hadoop-version": "2.6",
  "spark-worker-mem-mb": "10240"
  "yarn-master-mem-mb": "10240"
}
```

Config parameter for **jupyter** service type supports:
* **toree-version** -- use specific Toree version for Jupyter.

Example:
```json
"Config": {
  "toree-version": "1" 
}
```

Config parameter for **ignite** service type supports:
* **ignite-memory** -- percentage (integer number from 0 to 100) of worker memory to be assigned to Apache Ignite.
                       Currently this simply reduces spark executor memory, Apache Ignite memory usage must be manually configured.

Example:
```json
"Config": {
  "ignite-memory": "30" 
}
```

Config parameter for **elastic** service type supports:
* **es-heap-size** -- use specific ElasticSearch heap size. Default heap size is 1g (1 GB).

Example:
```json
"Config": {
  "es-heap-size": "1g" 
}
```

## Protobuf

Contains proto file for gRPC and have already generated code from it. Used in http_server and services/ansible_runner.

## http_server
Server that handles HTTP requests(probably from Envoy, that will take tham from real clients), and call, using gRPC(use client class from ansible-pb), ansible_service to lanch ansible.

# How to get it worked
Launch ansible_runner service:
```
go run src/services/ansible_service/ansible_service.go src/services/ansible_service/ansible_launch.go
```

Launch http_server:
```
go run src/http_server.go
```

Send request to localhost:8080/clusters":
```
curl localhost:8080/clusters -XPOST -d '{"Name":"spark-test","EntityStatus":1,"services":[{"Name":"spark-test","Type":"spark","Config":{"hadoop-version":"2.6", "use-yarn": "true"},"Version":"2.1.0"}],"NHosts":1}'
```

