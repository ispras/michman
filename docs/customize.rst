.. _michman_customize_section:

.. _project: https://github.com/ispras/spark-openstack

Modification of supported services
===================================

This section сovers how to customize Michmans services: add new versions, configuartion options and more. It also describes how to add new services to Michman. This guide will be useful to Michman's administrators and developers.


Ansible playbook description
-----------------------------

The Ansible-playbook used by the Michman is an extension and continuation of the Spark-Openstack `project`_. 

For "create", "update" and "delete" operations it's launched *main.yml* playbook.  In general, the playbook execution consists of the following steps:

	#. Creation of the cluster. Ansible runs role for launching VMs in the Cloud, based on requested cluster configurations, creates security group and gives floating IPs to created VMs. By default is used *async* mode for VMs creation. Then, user's ssh-key is deployed to the VMs. This step is executed on "create" operation.

	#. In the second stage are deployed some base packages and configurations that are listed in the *base* role. 

	#. Then services are deployed. Services are deployed on different groups of hosts. For now are supported three groups:
		* master
		* slave
		* storage

	 Ansible includes one or more Ansible-role/roles for every requested service. If variables used in service role are configured via Michman API, its should have name like following:

	 .. parsed-literal::
	 	<service_name>_<parameter_name>

	 If a choice of different versions is supported for a service, then such a variable must be defined for the version:

	 .. parsed-literal::
	 	<service_name>_version

	 If the role has dependencies, then they must be written in the "dependencies" section in the *meta*:

	 .. parsed-literal::
		---
		dependencies:
		 	- { role: spark_common }


Configuration of supported services
------------------------------------

You can modify or add following information in Michman's supported service (*service type*):
	* DefaultVersion -- service default version
	* Description -- service description
	* AccessPort -- default access port for service that will be writed to service URL
	* Ports -- you cant add information of service ports
	* Supported versions -- list of service versions

You should use Michman PUT */configs/{serviceType}* method for this action and provide JSON with modifications. Example:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/spark -XPUT -d 
	'{
		"AccessPort": 8080,
		"Ports": [
	      {
	        "Port": 8080,
	        "Description": "Spark GUI"
	      },
	      {
	        "Port": 50070,
	        "Description": "hdfs GUI"
	      }
    	]
	}'

You can modify or add following information in service version:
	* Description -- service version description
	* Configuration parameters -- service variables for this versions that could be customized by the user
	* Version dependencies from versions of other supported services

You should use Michman PUT */configs/{serviceType}/versions/{versionId}* method for this action and provide JSON with modifications. Example:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/spark -XPUT -d 
	'{
	      "Version":"9.6",
	      "Description":"PostgreSQL 9.6 version",
	      "Configs": [{
	          "ParameterName": "db_password",
	          "Type": "string",
	          "DefaultValue": "dbpassword",
	          "Required": true,
	          "Description": "Default user password for PostgreSQL DB for user postgres, you can change it"
	        }
	    
.. note:: You should also add abillity to deploy new service versions or its variables to ansible role for this service.

Adding a new service
---------------------

 In this section, we describe adding a new service to Michman using the IN-Memory DBMS Apache Ignite as an example. Service registration includes the following steps.

 	#. **Adding an Apache Ignite deployment Ansible role.** The role name should match the type of the service registered (in this case, ignite), all user-configurable variables for this role should have prefix in their name with the service type.

 	#. **Description of the registered service type in JSON format.** The service needs to describe information about supported versions, configurable parameters and dependencies. The class of service and the ability to access it are also indicated. The JSON example describes Apache Ignite service with supported version 7.1.1 and configurable working memory size is shown below. The Class field describes the connection between the service and the infrastructure. In the example, master-slave means that the service is deployed in a distributed form. For your convenience, we recommend adding this JSON document to the *init* directory.

	#. **Request to register a new system service.** The Michman administrator should send following request:

.. parsed-literal::
	curl -X POST -d "data=@michman/init/ignite.json" http://michman_addr:michman_port/configs

.. parsed-literal::
	
	#ignite service type definition
	{
	  "Type": "ignite",
	  "Description": "Apache Ignite service",
	  "DefaultVersion": "7.1.1",
	  "Class": "master-slave",
	  "Versions": [
	    {
	      "Version": "7.1.1",
	      "Description": "Apache Ignite default version for spark-openstack",
	      "Configs": [
	        {
	          "ParameterName": "memory",
	          "Type": "int",
	          "DefaultValue": "30",
	          "Required": true,
	          "Description": "percentage (integer number from 0 to 100) of worker memory to be assigned to Apache Ignite.\nCurrently this simply reduces spark executor memory, Apache Ignite memory usage must be manually configured."
	        }
	      ]
	    }]
	}

In case of a new service successful registration, a response returns to the user comprising HTTP-code 200 and JSON with a supplemented description of the "ignite" service type.
