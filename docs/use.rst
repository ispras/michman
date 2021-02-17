.. _michman_use_section:

Using Michman
=================

This guide сovers how to work with Michman and includes an simple example showing how to create a cluster in OpenStack cloud with a set of services.

Note: we use Michman without authentication (*use_auth: false*) for this example.

Create new project:

.. parsed-literal::
	curl {IP}:{PORT}/projects -XPOST -d '{"DisplayName":"test", "Description":"Project for tests", "DefaultImage": "centos"}'


Create new cluster with Jupyter and Spark services:

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "my-cluster",
		"Services": [
			{
				"Name": "jupyter-service",
				"Type": "jupyter"
			},
			{
				"Name": "spark-service",
				"Type": "spark",
				"Version": "2.3.0",
				"Configs": {
					"worker_mem_mb": "10240"
				}

			}
		],
		"Image": "ubuntu",
		"NHosts": 3
	}'


Get info about all clusters in project:

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters


Get info about **my-cluster** cluster in **test** project (**note: cluster name id constructed as cluster _DisplayName-ProjectName_**):

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters/my-cluster-test


Delete  **my-cluster** cluster in **test** project:

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters/my-cluster-test -XDELETE

Get service API in browser by this URL: **http://michman_addr:michman_port/api**

