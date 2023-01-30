![Michman](./docs/logo/michman-public-version-logo.png)

Michman is an orchestration self-hosted service intended to simplify process of creating distributed clusters and management of services in cloud environments. 
It provides capabilities for automatic deployment of distributed tools used in datascience, big data, HPC and business tasks:
* Apache Spark
* Apache Hadoop
* Apache Ignite
* Apache Cassandra
* ClickHouse
* CouchDB
* СVAT
* ElasticSearch with OpenDistro tools
* Greenplum
* Jupyter
* Jupyterhub
* Kubernetes
* MariaDB
* Nextcloud
* NFS-Server
* OpenPAI
* PostgreSQL
* Redis
* Slurm
* Distributed Tensorflow

More detailed description of supported services can be found at 
[docs/services.md](./docs/services.md).

Full documentation is available [here](https://michman.ispras.ru).

Clusters are created and managed via REST API (see swagger docs) with collaborative group-based access to computational resources.

This project follows up [spark-openstack project](https://github.com/ispras/spark-openstack) (ISP RAS).

## Quickstart
If you're familiar with Michman and already have access to Openstack and running Vault with required secrets and database (Couchbase or MySQL), there is the shortest way to run Michman tested on Ubuntu 20.04:
1. Clone Michman:
    ```shell
    git clone https://github.com/ispras/michman.git
    cd michman
    ```
   For deploying Kubernetes clusters git submodule must be installed:
    ```shell
    git submodule update --init --recursive
    ```
2. Fill at least Openstack, Vault and logs sections in the configuration file (look at `configs/config-sample.yaml`): 
3. Install go (skip if it's already installed):
    ```shell
    wget https://go.dev/dl/go1.18.8.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.18.8.linux-amd64.tar.gz
    echo "export PATH=$PATH:/usr/local/go/bin:/home/$USER/go/bin" >> ~/.profile
    source ~/.profile
    go version
    rm go1.18.8.linux-amd64.tar.gz
    ```
4. Install system packages:
    ```shell
    sudo apt update && sudo apt install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    git \
    protobuf-compiler \
    python3 \
    python3-pip \
    python3-setuptools \
    python3-venv \
    software-properties-common \
    unzip
    ```
5. Install go and python packages (virtualenv is highly recommended):
    ```shell
    go install github.com/golang/protobuf/protoc-gen-go
    go get github.com/go-sql-driver/mysql
    go get github.com/golang/mock/mockgen
    python3 -m venv ./venv
    source ./venv/bin/activate
    pip install --upgrade pip
    pip install --no-cache-dir -r requirements.txt
    ```
6. Compile and run (make sure virtualenv is activated, if it was configured on a previous step):
    ```shell
    ./build.sh start -c ./configs/config-sample.yaml # Replace with actual path
    ```
7. Check that Michman responds:
    ```shell
    curl localhost:8081/projects
    ```
Read [Advanced installation instructions](#Advanced installation instructions) for installation issues. 

Note, Michman requires information about images and flavors in Openstack and loaded descriptions of supported service types which are stored in `init` directory.

## Infrastructure requirements

* **Openstack** IaaS-provider. Supported versions: _Liberty_, _Stein_, _Ussuri_.
* Database server:
  * Last tested **Couchbase** version: 6.0.0 community edition. Couchbase must contain prepared buckets with primary indexes: _clusters_, _projects_, _templates_, _service_types_, _images_.Templates bucket is optional and is used if you are going to create templates.
  * Last tested **MySQL** version: 5.7 and **MariaDB** 10.3. Database should be created with sql/create_database.sql script.
* **Vault** server. Last tested version: 1.2.3

Read more about Michman configuration in corresponding [section](#Configuration).

## Configuration
### Openstack
Requirements:
* Images for virtual machines. LTS Ubuntu versions, CentOS 7,8 and Stream versions are available in Michman
* Floating ip pool
* Flavors depends on your proposes
* Key-pair

Note, images and flavors information should be stored in Michman (check Michman [usage section](#working-with-michman)).
You should write Openstack credentials and ssh private key corresponding to prepared key pair in Vault. Read [Vault configuration](#vault) section.

### Vault
Michman uses Vault for security issues.
It's required to create the following secrets:

1.**ssh_key**: it should contain field named `key_bgt` with private ssh_key corresponds to your Openstack key-pair in value.

2.**OpenStack**: general credentials are used to manage resources in Openstack. All required values must be in Openstack RC file. Depending on Openstack version this secret should contain the following keys (corresponding to environment variables in Openstack RC):
  * _Ussuri_:
    * `OS_AUTH_URL` &mdash; keystone address
    * `OS_IDENTITY_API_VERSION` &mdash; identity API version
    * `OS_INTERFACE` &mdash; interface type
    * `OS_PASSWORD` &mdash; authentication password
    * `OS_PROJECT_DOMAIN_ID` &mdash; domain ID containing project
    * `OS_PROJECT_ID` &mdash; OpenStack project-level authentication scope (by ID)
    * `OS_PROJECT_NAME` &mdash; OpenStack project-level authentication scope (by Name)
    * `OS_REGION_NAME` &mdash; authentication region name
    * `OS_USERNAME` &mdash; authentication username
    * `OS_USER_DOMAIN_NAME` &mdash; domain name or ID containing user
  * _Stein_:
    * `OS_AUTH_TYPE` &mdash; the authentication plugin type to use when connecting to the Identity service.
    * `OS_AUTH_URL` &mdash; usually means keystone address
    * `OS_PASSWORD` &mdash; authentication password
    * `OS_PROJECT_DOMAIN_NAME` &mdash; domain name containing project
    * `OS_PROJECT_NAME` &mdash; OpenStack project-level authentication scope (by Name)
    * `OS_REGION_NAME` &mdash; authentication region name
    * `OS_USERNAME` &mdash; authentication username
    * `OS_USER_DOMAIN_NAME` &mdash; domain name or ID containing user
  * _Liberty_:
    * `OS_AUTH_URL` &mdash; usually means keystone address
    * `OS_PASSWORD` &mdash; authentication password
    * `OS_PROJECT_NAME` &mdash; OpenStack project-level authentication scope (by Name)
    * `OS_REGION_NAME` &mdash; authentication region name
    * `OS_TENANT_ID` &mdash; authentication tenant ID
    * `OS_TENANT_NAME` &mdash; authentication tenant name
    * `OS_SWIFT_PASSWORD` (optional) &mdash; swift authentication password
    * `OS_SWIFT_USERNAME` (optional) &mdash; swift authentication username

    Depending on Openstack configuration additional parameters are required. Known examples for Stein OpenStack version: `COMPUTE_API_VERSION, NOVA_VERSION, OS_CLOUDNAME, OS_IMAGE_API_VERSION, OS_NO_CACHE, OS_VOLUME_API_VERSION, PYTHONWARNINGS, no_proxy`. You may contact your cloud admin to match required parameters.
    
    For now, we use these parameters directly in process environment for ansible, but clouds.yaml processing is coming soon.

3.**Database** credentials. Depending on database engine the secret should contain the following keys:
   * _Couchbase_:
     * `path` &mdash; Couchbase server address (ex. `127.0.0.1:8091`)
     * `password` &mdash; database user password
     * `username` &mdash; database username
   * _MySQL_:
     * `address` &mdash; MySQL server address (ex. `127.0.0.1:3306`)
     * `database` &mdash; database name
     * `password` &mdash; database user password
     * `user` &mdash; database username

4.**Hydra** secret (optional, it is used only for oauth2 authorization model) includes the following keys:
   * `redirect_uri` &mdash; OAuth 2.0 redirect URI
   * `client_id` &mdash; OAuth 2.0 client ID
   * `client_secret` &mdash; OAuth 2.0 client secret

5.**Docker registry** secret (optional) includes the following keys:

   * `url` &mdash; address of your selfsigned registry according to certificate or gitlab registry url

   * `user` &mdash; your selfsigned registry or gitlab registry username

   * `password` &mdash; your selfsigned registry or gitlab registry password


### Database
Now Michman may work with **Couchbase** and **MySQL** (or **MariaDB**).

[Couchbase](https://www.couchbase.com/) is json-based NoSQL DBMS with in-memory storage, horizontal scaling potential SQL-like query engine and other features.
Michman needs the following buckets with created [primary indexes](https://docs.couchbase.com/server/current/n1ql/n1ql-language-reference/createprimaryindex.html) to work with Couchbase:
* `clusters`: clusters created by Michman
* `flavors`: available Openstack flavors to run virtual machines
* `images`: available Openstack images to run virtual machines
* `projects`: Michman projects
* `service_types`: services available to deploy Michman
* `templates` (optional): templates of combined service types for easier deploy


[MySQL](https://www.mysql.com/) and [MariaDB](https://mariadb.org/) are similar traditional relational DBMS. Michman needs prepared database that may be created with `sql/create_tables.sql` script.

It's necessary to initialize Michman with supported Service Types stored in `init` directory. Read how to upload them in [Usage](#Usage) section.  

### Configuration file
There is template in `configs/config-sample.yaml`. You should fill at least the following common parameters:
* **os_key_name** &mdash; key pair name of your Openstack account
* **virtual_network** &mdash; OpenStack virtual network name or ID (in Neutron or Nova-networking)
* **floating_ip_pool** &mdash; Openstack floating IP pool name
* **os_version** &mdash; OpenStack version code name. For now supported next versions: Ussuri, Stein and Liberty
* **vault_addr** &mdash; Vault address
* **token** &mdash; Vault root token
* **os_key** &mdash; Vault path to Openstack credentials
* **ssh_key** &mdash; Vault path to secret with ssh private key
* **storage** &mdash; Type of used database. Acceptable values: _mysql_ or _couchbase_
* **cb_key** &mdash; Vault path to Couchbase credentials. Required if _couchbase_ storage is used
* **mysql_key** &mdash; Vault path to MySQL credentials. Required if _mysql_ storage is used
* **logs_output** &mdash; type of logging system. Acceptable values: _file_, _logstash_
* **logs_file_path** &mdash; path to directory with logs
* **logstash_addr** &mdash; logstash address if logstash output is used
* **elastic_addr** &mdash; elastic address if logstash output is used

These parameters should be filled to configure authorization:
* **use_auth** &mdash; boolean flag for authentication usage
* **authorization_model** &mdash; type of authorization model. Acceptable values: _none_, _oauth2_ or _keystone_
* **policy_path** &mdash; local path to policy configuration file (e.g. `configs/policy.csv`)
* **admin_group** &mdash; name of the admin group
* **session_idle_timeout** &mdash; time limit in minutes of a period, while session can be inactive before it expires
* **session_lifetime** &mdash; time limit in minutes of a period, while session is valid before it expires
* **hydra_key** &mdash; Vault path to hydra credentials. Required if _oauth2_ authorization model is used
* **hydra_admin** &mdash; address of hydra admin server. Required if _oauth2_ authorization model is used
* **hydra_client** &mdash; address of hydra client server. Required if _oauth2_ authorization model is used
* **keystone_addr** &mdash; address of keystone service. Required if _keystone_ authorization model is used

These parameters should be filled to configure usage of local repositories and registries:
* **use_package_mirror** &mdash; boolean flag for system package mirrors usage
* **use_pip_mirror** &mdash; boolean flag for pip mirror usage
* **apt_mirror_address** &mdash; address of Debian packages mirror
* **yum_mirror_address** &mdash; address of Redhat packages mirror
* **pip_mirror_address** &mdash; address of pip packages mirror
* **pip_trusted_host** &mdash; IP of used pip packages mirror
* **docker_insecure_registry** &mdash; boolean flag for local insecure (without certificates and user control) registry usage
* **docker_selfsigned_registry** &mdash; boolean flag for local self-signed registry usage
* **docker_gitlab_registry** &mdash; boolean flag for gitlab registry usage
* **docker_insecure_registry_ip** &mdash; host ip of your insecure registry
* **docker_selfsigned_registry_ip** &mdash; host ip of your selfsigned registry
* **docker_selfsigned_registry_port** &mdash; host port of your selfsigned registry
* **docker_selfsigned_registry_url** &mdash; address of your selfsigned registry according to certificate
* **docker_cert_path** &mdash; path to your selfsigned certificate
* **registry_key** &mdash; Vault path to gitlab registry credentials

## Usage
Michman provides REST API to interact with it (default used port is _8081_). OpenStack images and flavors descriptions, Michman project and service types have to be prepared before starting the process of clusters creation. If you have configured Michman using Keystone or OAuth2, you may also have to get authenticated to interact with Michman. Read more in full [documentation](https://michman.ispras.ru/en/use.html) or in Swagger after starting Michman on `localhost:8081/api`.
There are several basic examples of typical requests to Michman on localhost from _curl_:
* Write image information:
    ```bash
    curl -XPOST http://localhost:8081/images \
    --data '{
      "Name": "ubuntu20.04",
      "AnsibleUser": "ubuntu",
      "CloudImageID": "uuid-from-openstack"
    }'
    ```
* Write flavor information (for now _Name_ should be corresponding to Openstack flavor name):
    ```bash
    curl -XPOST http:localhost:8081/flavors \
    --data '{
      "Name": "Standard2.medium.s50",
      "VCPUs": 2,
      "RAM": 2048,
      "Disk": 50
    }'
    ```
* Write service type from json file:
    ```bash
    curl -XPOST http:localhost:8081/configs \
    --data @init/jupyter.json
    ```
* Create project:
    ```bash
    curl -XPOST http://localhost:8081/projects \
    --data '{
      "DisplayName": "readme-project",
      "DefaultImage": "ubuntu20.04",
      "Description": "Michman demo project",
      "DefaultMasterFlavor": "Standard2.medium.s50",
      "DefaultSlavesFlavor": "Standard2.medium.s50",
      "DefaultStorageFlavor":"Standard2.medium.s50",
      "DefaultMonitoringFlavor": "Standard2.medium.s50"
    }'
    ```
* Read project information:
    ```bash
    curl -XGET http://localhost:8081/projects/readme-project
    ```
* List all clusters of the project:
    ```bash
    curl -XGET http://localhost:8081/projects/readme-project/clusters
    ```
* Create new cluster with Jupyterlab service using default image and flavor from project:
    ```bash
    curl -XPOST http://localhost:8081/projects/readme-project/clusters \
      --data '{
        "DisplayName": "jupyter-cluster",
        "Services": [
            {
                "Name": "my-awesome-jupyterlab",
                "Type": "jupyter",
                "Version": "jupyter-lab"
            }
        ]
      }'
    ```

Get info about **jupyter-test** cluster in **Test** project (note that cluster _Name_ is constructed as cluster _DisplayName-ProjectName_):
```bash
curl localhost:8081/projects/readme/clusters/jupyter-test-readme
```

Delete  **jupyter-test** cluster in **Test** project:
```bash
curl localhost:8081/projects/readme/clusters/jupyter-test-readme -XDELETE
```

Get service API in browser by this URL: `localhost:8081/api`


## Advanced installation instructions
### Go version
For now _protoc-gen-go_ package is deprecated since go version 1.19. Moreover, _protobuf-compiler_ apt package wasn't available on Ubuntu versions before 18.04, may have no analogs for others distributions. If you use newer go versions or can't use _protobuf-compiler_ and _protoc-gen-go_ packages for other reason, you may try the following solutions:
1. The simplest way is to use docker _protoc_ image:
    ```shell
    docker pull znly/protoc
    # run from Michman directory
    ./build.sh -d proto # -d option enables docker image usage 
    ```
    (Docker installation instructions are [here](https://docs.docker.com/engine/install/). It is also recommended to configure [rootless mode](https://docs.docker.com/engine/install/linux-postinstall/))

    Note that docker container may generate files with permissions for root user, so check `internal/protobuf/launcher.pb.go` permissions.  
2. [Here](https://askubuntu.com/a/1072684) is the most stable installation for _libprotoc_ from source code.
3. You may try to use _protoc-gen-go_ package installed with go 1.18 or older to avoid deprecation alert.

### Python packages
Depending on your python version, some packages may be unavailable. For most of them you may just use the nearest to requirements.txt version.
Most common issue is related to `openstacksdk` version. There is [mapping](https://releases.openstack.org/teams/openstacksdk.html) of OpenStack versions and corresponding `openstacksdk` version. 

[Mitogen](https://github.com/mitogen-hq/mitogen) can be used for ansible acceleration. Note that it shouldn't be used in production now because the project is unstable.
To use Michman with Mitogen clone the instrument and add it to defaults section in ansible/ansible.cfg in source code of the orchestrator. 
Known issues with Mitogen:
* Mitogen enforces `StrictHostKeyChecking` on ssh connection for ansible;
* async tasks don't work;
* default ansible interpreter is replaced, so ansible fails when works with some images.

### Start up issues
`build.sh` script is used to manipulate with Michman. It includes commands to generate protobuf code and testing mocks, run tests and Michman as daemon, stop it, clean up directory and other. Check `./build.sh --help` command for more details.
Notes:
1. Script automatically runs protobuf code generation and compilation on start up command if it doesn't exist, so you may just run `./build.sh start`.
2. By default Michman reads settings from `configs/config.yaml` file. If you want to use configuration file from other place, use flag `-c`:
```
./build.sh -c ./my-michman-config.yaml start
```
3. There are two commands to clean up working directory: 
   * _clean_: removes all generated files (protobuf, mocks, binary etc.) if they exist
   * _reset_: runs _stop_ and _clean_ operations

If Michman doesn't start with `build.sh` script, you should check `.launch_start.log` and `.http_start.log` files.

Sometimes it's helpful to start Michman without script in development environment:
```bash
./build.sh proto # to generate protobuf code 
go run ./cmd/launcher/main.go &
go run ./cmd/rest/main.go &
```
In the case you should kill processes manually.