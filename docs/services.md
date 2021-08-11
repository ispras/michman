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
* Nextcloud
* NFS-Server
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
