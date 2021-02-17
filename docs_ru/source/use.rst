.. _michman_use_section:

Работа с Michman
=================

В этом разделе рассказывается, как работать с системой Michman, и демонстрируется простой пример, показывающий, как создать кластер в облаке OpenStack с набором из нескольких сервисов. 

Замечание: в наших примерах Michman используется без аутентификации (*use_auth: false*).

Создайте новый проект:

.. parsed-literal::
	curl {IP}:{PORT}/projects -XPOST -d '{"DisplayName":"test", "Description":"Project for tests", "DefaultImage": "centos"}'


Создайте в нем новый кластер с сервисами Jupyter и Spark:

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

Получите информацию обо всех созданных в проекте кластерах:

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters


Получите информацию о кластере **my-cluster** в проекте **test** (**замечание: имя кластер формируется как _DisplayName-ProjectName_**):

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters/my-cluster-test


Удалите кластер  **my-cluster** в проекте **test**:

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters/my-cluster-test -XDELETE

Получите информацию о Michman API в браузере по URL: **http://michman_addr:michman_port/api**

