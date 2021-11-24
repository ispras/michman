![Michman](./docs/logo/michman-public-version-logo.png)

Michman is an orchestration self-hosted service intended to simplify process of creating distributed clusters and management of services in cloud environments. At now it provides capabilities for deployment a part of Apache big data stack respecting user ability to choose needed versions with additional tools and services:
* Apache Spark
* Apache Hadoop
* Apache Ignite
* Apache Cassandra
* ClickHouse
* CouchDB
* СVAT
* ElasticSearch with OpenDistro tools
* Jupyter
* Jupyterhub
* Kubernetes
* Nextcloud
* NFS-Server
* Slurm
* PostgreSQL
* Redis

More detailed description of supported services can be found at 
[docs/services.md](./docs/services.md).

Clusters are created and managed via REST API (see swagger docs) with collaborative group-based access to computational resources.

This project follows up spark-openstack project (ISP RAS).

## Dependencies

Apt packages:
- golang
- libprotoc 3.6.1
- unzip
- apt-transport-https
- ca-certificates
- curl
- software-properties-common
- python, python-pip, python-setuptools

Python packages:
- ansible
- openstacksdk

Go packages:
- are listed in [go.mod](./go.mod) file
- protoc-gen-go (for generating grpc code)
- mockgen (for generating mocks)

## Infrastructure requirements

* **Openstack** cloud. Supported versions: _Liberty_, _Stein_, _Ussuri_.
  * Currently project supports deploying of services only on VMs with Ubuntu (16.04 or 18.04) or CentOS, so should be prepared suitable image.
  * It's recomended to prepare floating ip pool and flavors for created VMs.
  * Also you should prepare security key-pair and pem-key to provide access to created VMs from launcher. Key should be pasted in `$PROJECT_ROOT/ansible/files/ssh_key` file or in Vault secrets storage.
* **Couchbase** server.
  * Tested version: 6.0.0 community edition
  * Must contain prepared buckets with primary indexes: _clusters_, _projects_, _templates_, _service_types_, _images_. Templates bucket is optional and used only if you going to create templates.
  
* **Vault** server:
  * Tested version: 1.2.3
  * Stored secrets (Secret engine type - _kv v1_, path: kv/):
    * kv/couchbase: _path, username, password_, where path means address of Couchbase server
    * kv/openstack should contain authentication info from Openstack rc file: 
        * `OS_AUTH_URL, OS_PASSWORD, OS_PROJECT_NAME, OS_REGION_NAME, OS_TENANT_ID, OS_TENANT_NAME, OS_USERNAME` for Liberty Openstack version
        * `OS_AUTH_URL, OS_PASSWORD, OS_PROJECT_NAME, OS_REGION_NAME, OS_USERNAME, OS_SWIFT_USERNAME, OS_SWIFT_PASSWORD, COMPUTE_API_VERSION, NOVA_VERSION, OS_AUTH_TYPE, OS_CLOUDNAME, OS_IDENTITY_API_VERSION, OS_IMAGE_API_VERSION, OS_NO_CACHE, OS_PROJECT_DOMAIN_NAME, OS_USER_DOMAIN_NAME, OS_VOLUME_API_VERSION, PYTHONWARNINGS, no_proxy` for Stein Openstack version
        * `OS_AUTH_URL, OS_PASSWORD, OS_PROJECT_NAME, OS_PROJECT_ID, OS_REGION_NAME, OS_DOMAIN_ID, OS_INTERFACE, OS_USERNAME, OS_USER_DOMAIN_NANE, OS_IDENTITY_API_VERSION` for Ussurri Openstack version
    * kv/pem_key: _key_bgt_ - must contain actual ssh key from Openstack key pair for `OS_USERNAME`
* **Docker registry**  
  * Currently Nextcloud service deployment based on docker containers. It's possible to use local registry:
    1. Prepare your registry. It may be insecure registry (without any sertificates and user controls), selfsigned registry or gitlab registry.
    2. Configure:
      1. If you use insecure registry, set `docker_incecure_registry: true` and `insecure_registry_ip: xx.xx.xx.xx:xxxx` options in _config.yaml_
      2. If you use selfsigned registry, you need to set `docker_selfsigned_registry: true`, `docker_selfsigned_registry_ip: xx.xx.xx.xx:xxxx`, `docker_selfsigned_registry_url: consides.to.cert.url` and `docker_cert_path: path_to_registry_cert.crt` in _config.yaml_
      3. If you use gitlab registry you should set `docker_gitlab_registry: true`
    3. In case of using selfsigned or gitlab registry you should add secret with _url_, _user_ and _password_ to **vault** and set `registry_key: key_of_docker_secret` in _config.yaml_

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

Hydra (hydra_key) secret includes following key:
 * **redirect_uri** -- OAuth 2.0 redirect URI
 * **client_id** -- OAuth 2.0 client ID
 * **client_secret** -- OAuth 2.0 client secret
  
This secret is optional, it is used only for oauth2 authorization model.

## Configuration

Configuration of the project is stored in the **configs/config.yaml** file. Example:

```yaml
## Vault
token: MY-TOKEN
vault_addr: http://127.0.0.1:8200
os_key: kv/openstack
ssh_key: kv/ssh-keys
cb_key: kv/couchbase/
hydra_key: kv/hydra

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
mirror_address: ip

## Docker registry
docker_insecure_registry: true
docker_selfsigned_registry: true
docker_gitlab_registry: false

docker_insecure_registry_ip: ip:5000
docker_selfsigned_registry_ip: ip
docker_selfsigned_registry_url: bgtregistry.ru
docker_cert_path: /home/ubuntu/docker.crt 

#auth
use_auth: true
authorization_model: none 
admin_group: admin
session_idle_timeout: 480 
session_lifetime: 960 

#hydra auth params
hydra_admin: HYDRA_ADDR
hydra_client: HYDRA_ADDR

#keystone params
keystone_addr: KEYSTONE_ADDR

#cluster logs
logs_output: file #file or logstash
logs_file_path: /home/ubuntu/go/src/gitlab.at.ispras.ru/michman/logs
logstash_addr: http://ip:9000
elastic_addr: http://ip:9200
```

Where:
* **token** - your token for authorization in vault 
* **vault_addr** - address of you vault db
* **os_key** - path to openstack's kv secrets engine
* **ssh_key** - path to ssh_key's kv secrets engine
* **cb_key** - path to couchbase's kv secrets engine
* **hydra_key** -- path to hydra's kv secrets engine
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
* **use_auth** -- use authentication with sessions or no
* **authorization_model** -- supports none, oauth2 or keystone values
* **admin_group** -- name of the admin group
* **session_idle_timeout** -- time in minutes, controls the maximum length of time a session can be inactive before it expires
* **session_lifetime** -- time in minutes, controls the maximum length of time that a session is valid for before it expires
* **hydra_admin** -- address of hydra admin server, if oauth2 authorization model is used
* **hydra_client** -- address of hydra client server, if oauth2 authorization model is used
* **keystone_addr** -- address of keystone service, if keystone authorization model is used
* **logs_output** -- output for cluster deployment logs, supports file or logstash values
* **logs_file_path** -- path to directory for cluster deployment logs if file output is used
* **logstash_addr** -- logstash address if logstash output is used
* **elastic_addr** -- elastic address if logstash output is used

## Getting started

### Install dependencies

Before start, make sure, all the dependencies are installed and go environment
is set up.

```bash
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
sudo apt install golang-go
sudo apt install unzip apt-transport-https \
  ca-certificates curl software-properties-common \
  python python-pip python-setuptools
```

Also, `libprotoc 3.6.1` is required. Working installation discribed [here](https://askubuntu.com/questions/1072683/how-can-i-install-protoc-on-ubuntu-16-04) or may be used docker container [like this](https://hub.docker.com/r/znly/protoc/):
```bash
docker pull znly/protoc
```

Python packages:
```bash
pip install -r requirements.txt
# or
pip3 install -r requirements.txt
```

Go packages:
```bash
go get -u github.com/golang/protobuf/protoc-gen-go
go get github.com/golang/mock/mockgen
```

### Build

First, place the project code in the $GOPATH:
```bash
mkdir -p $GOPATH/src/github.com/ispras
git clone https://github.com/ispras/michman.git $GOPATH/src/github.com/ispras/michman
  
cd $GOPATH/src/github.com/ispras/michman
git submodule update --init --recursive
```

Then, complete _config.yaml_ file. Note: we use Michman without authentication 
(use_auth: false) for this example.

> Note: if you use protoc inside a docker container, set `USE_DOCKER` 
> environment variable to `true` before running `build.sh` script.

To quick start you may use [build.sh](./build.sh) script:
```bash
./build.sh start
```

Or services can be launched manually:

### ansible_runner

Manually launch ansible_runner service:
```bash
go run ./cmd/launcher
```

Manually launch ansible_runner service specifying config and port, defaults are config path in Michman root and 5000 as used port:
```bash
go run ./cmd/launcher --config /path/to/config.yaml --port PORT
```

### api_server

Manually launch api_server:
```bash
go run ./cmd/rest
```

Manually launch api_server specifying config, port and launcher address, defaults are config path in Michman root, 8081 as used port and localhost:5000 for launcher address:
```bash
go run ./cmd/rest --config /path/to/config.yaml --port PORT --launcher launcher_host:launcher_port
```

## Working with Michman

Clusters are created and managed via HTTP requests to the api_server.

Create new project:
```bash
curl {IP}:{PORT}/projects -XPOST -d '{
  "Name":"Test", 
  "Description":"Project for tests"
}'
```

Create new cluster with Jupyter service:
```bash
curl {IP}:{PORT}/projects/{ProjectID}/clusters -XPOST -d '{
  "DisplayName":"jupyter-test", 
  "Services":[
    {
      "Name":"jupyter-project",
      "Type":"jupyter"
    }
  ],
  "NHosts":1
}'
```

Get info about all clusters in project:
```bash
curl {IP}:{PORT}/projects/{ProjectID}/clusters
```

Get info about **jupyter-test** cluster in **Test** project (**note: cluster name id constructed as cluster _DisplayName-ProjectName_**):
```bash
curl {IP}:{PORT}/projects/{ProjectID}/clusters/jupyter-test-Test
```

Delete  **jupyter-test** cluster in **Test** project:
```bash
curl {IP}:{PORT}/projects/{ProjectID}/clusters/jupyter-test-Test -XDELETE
```

Get service API in browser by this URL: `{IP}:{PORT}/api`

