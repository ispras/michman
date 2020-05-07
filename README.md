![Michman](michman-public-version-logo.png)

Michman is an orchestration self-hosted service intended to simplify process of creating distributed clusters and management of services in cloud environments. At now it provides capabilities for deployment a part of Apache big data stack respecting user ability to choose needed versions with additional tools and services:
* Apache Spark
* Apache Hadoop
* Apache Ignite
* Apache Cassandra
* ElasticSearch with OpenDistro tools
* Jupyter
* Jupyterhub
* Nextcloud

Clusters are created and managed via REST API (see swagger docs) with collaborative group-based access to computational resources.

This project follows up spark-openstack project (ISP RAS).

## Dependencies
Apt packages:
```shell script
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
sudo apt install golang-go
sudo apt install unzip apt-transport-https \
  ca-certificates curl software-properties-common \
  python python-pip python-setuptools
```

Python packages:
```shell script
pip install ansible==2.9.4 openstacksdk==0.40.0 # latest tested versions or
# pip3 install ansible==2.9.4 openstacksdk==0.40.0
```
Go packages:
```shell script
go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
go get -u github.com/hashicorp/vault/api
go get -u gopkg.in/yaml.v2
go get github.com/julienschmidt/httprouter
go get -u gopkg.in/couchbase/gocb.v1
go get github.com/google/uuid
go get github.com/golang/mock/gomock
go get github.com/golang/mock/mockgen
```

Also required `libprotoc 3.6.1`. Working installation discribed [here](https://askubuntu.com/questions/1072683/how-can-i-install-protoc-on-ubuntu-16-04) or may be used docker container [like this](https://hub.docker.com/r/znly/protoc/). Example:
```
docker pull znly/protoc
cd $GOPATH/src/github.com/ispras/michman/protobuf
docker run --rm -v $(pwd):$(pwd) -w $(pwd) znly/protoc --go_out=plugins=grpc:. -I. protofile.proto
```
## Infrastructure requirements
* **Openstack** cloud. Supported versions: _Liberty_, _Stein_.
  * Currently project supports deploying of services only on VMs with Ubuntu (16.04 or 18.04), so should be prepared suitable image.
  * It's recomended to prepare floating ip pool and flavors for created VMs.
  * Also you should prepare security key-pair and pem-key to provide access to created VMs from launcher. Key should be pasted in `$PROJECT_ROOT/launcher/ansible/files/ssh_key` file.
* **Couchbase** server.
  * Tested version: 6.0.0 community edition
  * Must contain prepared buckets with primary indexes: _clusters_, _projects_, _templates_. The last one is optional and used only if you going to create templates.
* **Vault** server:
  * Tested version: 1.2.3
  * Stored secrets (Secret engine type - _kv v1_, path: kv/):
    * kv/couchbase: _path, username, password_, where path means address of Couchbase server
    * kv/openstack: 
        -- `OS_AUTH_URL, OS_PASSWORD, OS_PROJECT_NAME, OS_REGION_NAME, OS_TENANT_ID, OS_TENANT_NAME, OS_USERNAME` for Liberty Openstack version
        -- `OS_AUTH_URL, OS_PASSWORD, OS_PROJECT_NAME, OS_REGION_NAME, OS_USERNAME, OS_SWIFT_USERNAME, OS_SWIFT_PASSWORD, COMPUTE_API_VERSION, NOVA_VERSION, OS_AUTH_TYPE, OS_CLOUDNAME, OS_IDENTITY_API_VERSION, OS_IMAGE_API_VERSION, OS_NO_CACHE, OS_PROJECT_DOMAIN_NAME, OS_USER_DOMAIN_NAME, OS_VOLUME_API_VERSION, PYTHONWARNINGS, no_proxy` for Stein Openstack version
    * kv/pem_key: _key_bgt_ - must contain actual ssh key from Openstack key pair for `OS_USERNAME`
* **Docker registry**  
  * Currently Nextcloud service deployment based on docker containers. It's possible to use local registry:
    1. Prepare your registry. It may be insecure registry (without any sertificates and user controls), selfsigned registry or gitlab registry.
    2. Configure:
      1. If you use insecure registry, set `docker_incecure_registry: true` and `insecure_registry_ip: xx.xx.xx.xx:xxxx` options in _config.yaml_
      2. If you use selfsigned registry, you need to set `docker_selfsigned_registry: true`, `docker_selfsigned_registry_ip: xx.xx.xx.xx:xxxx`, `docker_selfsigned_registry_url: consides.to.cert.url` and `docker_cert_path: path_to_registry_cert.crt` in _config.yaml_
      3. If you use gitlab registry you should set `docker_gitlab_registry: true`
    3. In case of using selfsigned or gitlab registry you should add secret with _url_, _user_ and _password_ to **vault** and set `registry_key: key_of_docker_secret` in _config.yaml_

## Configurations
Configuration of the project is stored in **config.yaml** file. Example:
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
os_version: liberty

## Apt and Pip mirror
use_mirror: false
mirror_address: 10.10.11.111

## Docker registry
docker_insecure_registry: true
docker_selfsigned_registry: true
docker_gitlab_registry: false

docker_insecure_registry_ip: 10.10.17.242:5000
docker_selfsigned_registry_ip: 10.10.17.246
docker_selfsigned_registry_url: bgtregistry.ru
docker_cert_path: /home/ubuntu/docker.crt 
```

Where:
* **token** - your token for authorization in vault 
* **vault_addr** - address of you vault db
* **os_key** - path to openstack's kv secrets engine
* **ssh_key** - path to ssh_key's kv secrets engine
* **cb_key** - path to couchbase's kv secrets engine
* **os_key_name** - key pair name of your Openstack account
* **virtual_network** - your virtual network name or ID (in Neutron or Nova-networking)
* **os_image** - ID of OS image
* **floating_ip_pool** - floating IP pool name
* **\<type\>_flavor** - instance flavor that exists in your Openstack environment (e.g. spark.large). Master and slaves flavors are required parameters anyway. Storage flavor required if you want deploy nextcloud or nfs-server 
* **os_version** - OpenStack version code name. **_Now are supported only two versions: "stein" and "liberty"_**
* **use_mirror** - Do or do not use your apt and pip mirror (optional)
* **mirror_address** - Address of you mirror. Can be omitted if use_mirror is false (optional)
* **docker_insecure_registry** - Do or not use local insecure (without certificates and user control) registry (optional)
* **docker_selfsigned_registry** - Do or not use local selfsigned registry (optional)
* **docker_gitlab_registry** - Do or not use gitlab registry (optional)
* **docker_insecure_registry_ip** - Host ip of your insecure registry (optional)
* **docker_selfsigned_registry_ip** - Host ip of your selfsigned registry (optional)
* **docker_selfsigned_registry_url** - Address of your selfsigned registry according to certificate (optional)
* **docker_cert_path** - path to your selfsigned certificate (optional)

## Vault secrets 

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

Docker registry (registry_key) secret includes following key:
 * **url** -- Address of your selfsigned registry according to certificate or gitlab registry url
 * **user** -- Your selfsigned registry or gitlab username
 * **password** -- Your selfsigned registry or gitlab password
 
 This secret is optional.

# Services

Supported services types are:
* **cassandra**
* **spark**
* **elastic**
* **jupyter**
* **ignite**
* **jupyterhub** 
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
  "spark-worker-mem-mb": "10240",
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
* **heap-size** -- use specific ElasticSearch heap size. Default heap size is 1g (1 GB).

Example:
```json
"Config": {
  "heap-size": "1g" 
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

* **weblab_name** -- name of Web Laboratory.
* **nfs_server_ip** -- NFS server IP.
* **mariadb_image** -- your docker image with mariadb
* **nextcloud_image** -- your docker image with nextcloud

Example:
```json
"Config": {
  "weblab_name": "Name",
  "nfs_server_ip": "IP",
  "mariadb_image": "bgtregistry.ru:5000/mariadb",
  "nextcloud_image": "bgtregistry.ru:5000/nextcloud"
}
```

## Protobuf

Contains proto file for gRPC that must be used to generate protobuf package. Used in rest and launcher services.

# How to get it worked
First, place project code in $GOPATH:
```shell script
git clone https://github.com/ispras/michman.git
mkdir $GOPATH/src/github.com
mkdir $GOPATH/src/github.com/ispras
mv ./michman $GOPATH/src/github.com/ispras/
cd $GOPATH/src/github.com/ispras/michman
```
Then, complete _config.yaml_ file.

To quick start you may use _build.sh_ script:
```
./build.sh start
```

Manually launch ansible_runner service:
```
go run ./launcher/ansible_launcher.go ./launcher/main.go
```

Manually launch ansible_runner service specifying config:
```
go run ./launcher/ansible_launcher.go ./launcher/main.go /path/to/config.yaml
```

Manually launch http_server:
```
go run ./rest/main.go
```

Manually launch http_server specifying config:
```
go run ./rest/main.go /path/to/config.yaml
```

Create new project:
```
curl {IP}:8080/projects -XPOST -d '{"Name":"Test", "Description":"Project for tests"}'
```

Create new cluster with Jupyter service:
```
curl {IP}:8080/projects/{ProjectID}/clusters -XPOST -d '{"DisplayName":"jupyter-test", "Services":[{"Name":"jupyter-project","Type":"jupyter"}],"NHosts":1}'
```

Get info about all clusters in project:
```
curl {IP}:8080/projects/{ProjectID}/clusters
```

Get info about **jupyter-test** cluster in **Test** project (**note: cluster name id constructed as cluster _DisplayName-ProjectName_**):
```
curl {IP}:8080/projects/{ProjectID}/clusters/jupyter-test-Test
```

Delete  **jupyter-test** cluster in **Test** project:
```
curl {IP}:8080/projects/{ProjectID}/clusters/jupyter-test-Test -XDELETE
```

Get service API in browser by this URL: **{IP}:8080/api**
