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

HTTP routing:
```
go get github.com/julienschmidt/httprouter
```

Couchbase SDK:
```
go get -u gopkg.in/couchbase/gocb.v1
```

UUIDs:
```
go get github.com/google/uuid
```

Ansible version >= 2.8.1 

## Configurations
Configuration of the service is stored in **config.yaml** file. Example:
```yaml
## Vault
token: MY-TOKEN
vault_addr: http://127.0.0.1:8200
os_key: kv/openstack
ssh_key: kv/ssh-keys
cb_key: kv/couchbase/

#Openstack
os_key_name: my_key
virtual_network: test
os_image: 731c8c7d-47fd-4b69-bdb4-00415e3ccb00
floating_ip_pool: test
master_flavor: keystone.medium
slaves_flavor: keystone.medium
storage_flavor: keystone.medium
fanlight_flavor: keystone.medium
os_version: liberty

## Apt and Pip mirror
use_mirror: false
mirror_address: 10.10.11.111
```


Where:
* **token** - your token for authorization in vault 
* **vault_addr** - address of you vault db
* **os_key** - path to openstack's kv secrets engine
* **ssh_key** - path to ssh_key's kv secrets engine
* **cb_key** - path to couchbase's kv secrets engine
* **os_key_name** - key pair name
* **virtual_network** - your virtual network name or ID (in Neutron or Nova-networking)
* **os_image** - ID of OS image
* **floating_ip_pool** - floating IP pool name
* **\<type\>_flavor** - instance flavor that exists in your Openstack environment (e.g. spark.large). Master and slaves flavors are required parameters anyway. Fanlight flavor required if you want deploy fanlight. Storage flavor required if you want deploy nextcloud or nfs-server 
* **os_version** - OpenStack version code name. **_Now are supported only two versions: "stein" and "liberty"_**
* **use_mirror** - Do or do not use your apt and pip mirror
* **mirror_address** - Address of you mirror. Can be omitted if use_mirror is false
## Services


Openstack (os_key) secrets includes following keys for **Liberty** version:
* **OS_AUTH_URL**
* **OS_PASSWORD**
* **OS_PROJECT_NAME**
* **OS_REGION_NAME**
* **OS_TENANT_ID**
* **OS_TENANT_NAME**
* **OS_USERNAME** 
* **OS_SWIFT_USERNAME** -- optional
* **OS_SWIFT_PASSWORD** -- optional 

Openstack (os_key) secrets includes following keys for **Stein** version:
* **OS_AUTH_URL**
* **OS_PASSWORD**
* **OS_PROJECT_NAME**
* **OS_REGION_NAME**
* **OS_USERNAME** 
* **COMPUTE_API_VERSION**
* **NOVA_VERSION**
* **OS_AUTH_TYPE**
* **OS_CLOUDNAME**
* **OS_IDENTITY_API_VERSION**
* **OS_IMAGE_API_VERSION**
* **OS_NO_CACHE**
* **OS_PROJECT_DOMAIN_NAME**
* **OS_USER_DOMAIN_NAME**
* **OS_VOLUME_API_VERSION**
* **PYTHONWARNINGS**
* **no_proxy**

Ssh (ssh_key) secrets includes following keys:
* **key_bgt** -- private ssh key for Ansible commands

Couchbase (cb_key) secretes includes following keys:
* **clusterBucket** -- name of the bucket storing clusters
* **password** -- password of couchbase
* **path** -- address of couchbase
* **username** -- user name of couchbase 

Contains service for ansible launching.

Supported services types are:
* **cassandra**
* **spark**
* **elastic**
* **jupyter**
* **ignite**
* **jupyterhub** 
* **fanlight**
* **nfs-server**
* **nextcloud**

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

Config parameter for **fanlight** service type supports:
* **fanlight_instance_url** -- Fanlight frontend vnclient base URL to build site links:
```
<link rel="stylesheet" href="fanlight_instance_url/vnc-toolbar/vnc-toolbar.min.css">
```
* **desktop_access_url** -- This parameter (without protocol part and desktop id at the end)
                            is used to construct noVNC rfb connection URL on the web page:
```
(ws|wss)://DesktopAccessURL/desktop_id
```

* **users_add** -- adds users on startup.
* **apps_add** -- list of applications to add on startup.
* **weblab_name** -- name of Web Laboratory.
* **fileshare_ui_ip** 
* **nfs_server_ip** -- NFS server IP.

Example:
```json
"Config": {
  "fanlight_instance_url": "https://mydomain/myservice/",
  "desktop_access_url": "https://mydomain/myservice/",
  "users_add": '[{"uuid":"", "pos_name":"", "pos_group":"", "pos_uid": 123, "pos_gid": 123}]'
  "apps_add": '[{"uuid":"myapp_uuid", "image":"docker_image_url"}]',
  "weblab_name": "Name",
  "fileshare_ui_ip": "IP",
  "nfs_server_ip": "IP"
}
```

Config parameter for **nfs-server** service type supports:
* **weblab_name** -- name of Web Laboratory.

Example:
```json
"Config": {
  "weblab_name": "Name"
}
```

Config parameter for **nextcloud** service type supports:

* **custom_oidc_providers_host** --  keycloak hostname.
* **custom_oidc_providers_ip** -- keycloak server ip.
* **nextcloud_url** -- URL of nextcloud as it will be opened from portal (including proxy and weblab id).
* **weblab_name** -- name of Web Laboratory.
* **nfs_server_ip** -- NFS server IP.

Example:
```json
"Config": {
  "custom_oidc_providers_host": "auth.sci-portal.gov.ru",
  "custom_oidc_providers_ip": "10.10.16.51",
  "nextcloud_url": "web.sci-portal.gov.ru/proxy/nfs-lab10",
  "weblab_name": "Name",
  "nfs_server_ip": "IP"
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

Launch ansible_runner service specifying config:
```
go run src/services/ansible_service/ansible_service.go src/services/ansible_service/ansible_launch.go /path/to/config.yaml
```

Launch http_server:
```
go run src/http_server.go
```

Launch http_server specifying config:
```
go run src/http_server.go /path/to/config.yaml
```

Create new project:
```
curl localhost:8080/clusters -XPOST -d '{"Name":"spark-test", "Services":[{"Name":"spark-test","Type":"spark","Config":{"hadoop-version":"2.6", "use-yarn": "true"},"Version":"2.1.0"}],"NHosts":1}'
```