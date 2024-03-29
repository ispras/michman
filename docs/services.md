# Services

Supported services types are:
* Apache Spark
* Apache Hadoop
* Apache Ignite
* Apache Cassandra
* ClickHouse
* CouchDB
* ElasticSearch with OpenDistro tools
* Jupyter
* Jupyterhub
* Kubernetes
* Nextcloud
* NFS-Server
* OpenPAI
* Slurm
* PostgreSQL
* Redis

## Spark

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

## Slurm 
Config parameter for **slurm** service type supports:
* **use_rest** -- parameter for setting or not Slurm REST API.
* **db_password** -- password for database.
* **db_user** -- user for database.
* **TaskPluginParam** -- parameter of configuration file slurm.conf. More detailed information about the parameter can be found at [Slurm docs](https://slurm.schedmd.com/slurm.conf.html).  
* **use_open_foam** -- parameter for using or not OpenFOAM with Slurm.
* **config_dir** -- path to template of configuration file slurm.conf.
* **cgroup_config_dir** -- path to template of configuration file cgroup.conf.
* **use_open_mpi** -- parameter for using or not OpenMPI with Slurm.
* **partitions** -- list that describes partitions of Slurm-cluster. More detailed information about the parameter can be found at [Slurm docs](https://slurm.schedmd.com/documentation.html)
* **open_mpi_version** -- version of OpenMPI.

Example:
```json
"Config": {
  "use_rest": "true",
  "db_password": "password",
  "db_user": "user",
  "TaskPluginParam": "Cores",
  "use_open_foam": "true",
  "config_dir": "templates/slurm/slurm.conf.j2",
  "cgroup_config_dir": "templates/slurm/cgroup.conf.j2",
  "use_open_mpi": "true",
  "partitions": "main:5",
  "open_mpi_version":"v1.10"
}
```

## Jupyter

Config parameter for **jupyter** service type supports:
* **toree-version** -- use specific Toree version for Jupyter.

Example:
```json
"Config": {
  "toree-version": "1" 
}
```

## Ignite

Config parameter for **ignite** service type supports:
* **ignite-memory** -- percentage (integer number from 0 to 100) of worker memory to be assigned to Apache Ignite.
                       Currently this simply reduces spark executor memory, Apache Ignite memory usage must be manually configured.

Example:
```json
"Config": {
  "ignite-memory": "30" 
}
```

## ElasticSearch

Config parameter for **elastic** service type supports:
* **heap-size** -- use specific ElasticSearch heap size. Default heap size is 1g (1 GB).

Example:
```json
"Config": {
  "heap-size": "1g" 
}
```

## Kubernetes

Config parameters for **Kubernetes** service type are:
* **network_plugin** -- CNI plugin responsible for configuring overlay network.

  Available options:
    - calico
    - flannel
    - weave
  Default: calico

* **container_runtime** -- container runtime environment in which all the
  containers are deployed.

  Available options:
    - docker
    - containerd
    - cri-o
  Default: docker

  If container runtime implements CRI, [crictl](https://kubernetes.io/docs/tasks/debug-application-cluster/crictl/)
  is installed.

* **enable_dashboard** indicates if Kubernetes dashboard UI will be installed.

  By default, dashboard is not exposed to the outer world, as it can be done
  [in many different ways](https://github.com/kubernetes/dashboard/blob/master/docs/user/accessing-dashboard/README.md).

* **enable_netchecker** indicates if the [netchecker](https://github.com/Mirantis/k8s-netchecker-server)
  service will be deployed.

  To get the most recent and cluster-wide network connectivity report, run
  from any of the cluster nodes:

  ```bash
  curl http://<netchecker_service_ip>:31081/api/v1/connectivity_check
  ```

* **enable_helm** indicates if [helm](https://helm.sh/) package manager will
  be installed.

* **enable_ingress_nginx** indicates if nginx ingress controller will be
  deployed.

  Check with
  ```bash
  kubectl get all -n ingress-nginx
  ```
* **enable_cinder_csi** indicates if Cinder CSI plugin will be installed. This
  allows one to request persistent storage for pods right from the kubernetes.

* **enable_keystone_auth** indicates if Keystone Webhook authentication and
  authorization is available.

  To access kubernetes cluster via `kubectl` you need
  * Download latest `client-keystone-auth` from [GitHub](https://github.com/kubernetes/cloud-provider-openstack/releases)
  * Configure your user in `KUBECONFIG` to use this client to get a token from Kubernetes
    ```yaml
    - name: your-user
      user:
        exec:
          command: path-to-client-keystone-auth
          apiVersion: client.authentication.k8s.io/v1beta1
    ```

  More details in [k8s-keystone-auth](https://github.com/kubernetes/cloud-provider-openstack/blob/master/docs/keystone-auth/using-keystone-webhook-authenticator-and-authorizer.md) documentation.

## NFS-Server

Config parameter for **nfs-server** service type supports:
* **weblab_name** -- name of Web Laboratory.

Example:
```json
"Config": {
  "weblab_name": "Name"
}
```

## Nextcloud

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

## OpenPAI

Config parameter for **openpai** service type supports:

* **admin_username** -- name for the admin user.
* **admin_password** -- password for the admin user.

For this service, the kubernetes **container_runtime** must be left by default (docker).

Example:
```json
"Config": {
  "admin_username": "michman",
  "admin_password": "michman-pswd"
}
```
