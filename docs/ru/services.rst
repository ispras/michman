.. _michman_services_section:

.. _Ignite: https://ignite.apache.org

.. _Cassandra: https://cassandra.apache.org

.. _ClickHouse: https://clickhouse.tech

.. _CouchDB: https://couchdb.apache.org

.. _PostgreSQL: https://www.postgresql.org

.. _Redis: https://redis.io

.. _Spark: https://spark.apache.org

.. _Hadoop: https://hadoop.apache.org

.. _YARN: https://spark.apache.org/docs/latest/running-on-yarn.html

.. _JupyterLab: https://jupyter.org

.. _JupyterHub: https://jupyterhub.readthedocs.io/en/stable/

.. _Nextcloud: https://nextcloud.com

.. _Elasticsearch: https://www.elastic.co

.. _Kubernetes: https://kubernetes.io

.. _Slurm: https://slurm.schedmd.com/documentation.html

.. _MariaDB: https://mariadb.org/

Поддерживаемые сервисы
=======================

В этом разделе представлена информация о сервисах, которые можно развернуть с помощью Michman. 

.. image:: _static/Services.png

Облачные СУБД
-----------------

Michman поддерживает набор различных СУБД, которые можно легко развернуть в облаке. 

**Apache Ignite**

Apache `Ignite`_ это распределенная база данных для высокопроизводительных вычислений в оперативной памяти. Эта СУБД может быть развернута в распределенном кластере виртуальных машин в облаке с помощью Michman.

На текущий момент в системе поддерживается версия *7.1.1* Apache Ignite. 

Параметр конфигурации сервиса **ignite** включает:

* **memory** -- процент (целое число от 0 до 100) рабочей памяти, которая будет назначена Apache Ignite.

Следующий пример показывает запрос для создания кластера из трех узлов с сервисом Apache Ignite и уточненным параметром *memory*.

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "my-cluster",
		"Services": [
			{
				"Name": "ignite-service",
				"Type": "ignite",
				"Version": "7.1.1",
				"Configs": {
					"memory": "30"
				}

			}
		],
		"Image": "ubuntu",
		"NHosts": 3
	}'

Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса Apache Ignite доступна по API:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/ignite

**Apache Cassandra**

Apache `Cassandra`_ это распределенная система управления базами данных NoSQL с открытым исходным кодом, относящаяся к колоночным СУБД, предназначенная для обработки больших объемов данных, обеспечивая высокую доступность без единой точки отказа. Этот сервис можно развернуть в распределенном кластере виртуальных машин в облаке с помощью Michman. 

Michman позволяет развернуть Apache Cassandra, связанную с системой Spark.

На текущий момент поддерживается версия *3.11.4* Apache Cassandra. 

Следующий пример показывает запрос для создания кластера из трех узлов с сервисами Apache Cassandra и Spark.

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "my-cluster",
		"Services": [
			{
				"Name": "cassandra-service",
				"Type": "cassandra",
				"Version": "3.11.4"
			},
			{
				"Name": "spark-service",
				"Type": "spark",
				"Version": "2.3.0",
			}
		],
		"Image": "ubuntu",
		"NHosts": 3
	}'

Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса Apache Cassandra доступна по API: 

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/cassandra

**ClickHouse**

`ClickHouse`_ это высокопроизводительная система управления базами данных OLAP с открытым исходным кодом. Этот сервис можно было развернуть на узле хранения в облаке с помощью Michman. 

На текущий момент поддерживается версия *latest* ClickHouse. 

Config parameter for **clickhouse** service type supports:

* **db_password** -- Default user password for Clickhouse DB for user 'default', you can change it.


Следующий пример показывает запрос для развертывания в облаке СУБД ClichHouse.

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "my-cluster",
		"Services": [
			{
				"Name": "clickhouse-service",
				"Type": "clickhouse",
				"Configs": {
					"db_password": "secret"
				}
			}
		],
		"Image": "ubuntu",
		"NHosts": 1
	}'

Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса ClickHouse доступна по API:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/clickhouse

**Apache CouchDB**

Apache `CouchDB`_ это документо-ориентированная база данных NoSQL с открытым исходным кодом, реализованная на Erlang. Этот сервис можно было развернуть на узле хранения в облаке с помощью Michman. 

На текущий момент поддерживается версия *latest* CouchDB. 

Параметр конфигурации сервиса **couchdb** включает:

* **db_password** -- Пароль пользователя по умолчанию для CouchDB для пользователя 'admin', вы можете его изменить позже.


Следующий пример показывает запрос для развертывания в облаке СУБД CouchDB.

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "my-cluster",
		"Services": [
			{
				"Name": "couchdb-service",
				"Type": "couchdb",
				"Configs": {
					"db_password": "secret"
				}
			}
		],
		"Image": "ubuntu",
		"NHosts": 1
	}'

Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса CouchDB доступна по API:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/couchdb

**MariaDB**

`MariaDB`_ - реляционная база данных с открытым исходным кодом. При помощи системы оркестрации Michman данный сервис может быть развернут в качетстве облачного хранилища.
Конфигурационные параметры сервиса:

* **db_password** -- пароль для базы данных. Значение по умолчанию: password.
* **db_user** -- пользователь базы данных. Значение по умолчанию: user. 

Следующий пример показывает запрос для развертывания в облаке СУБД MariaDB.

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "test",
		"Services": [
			{
				"Name": "mariadb",
				"Type": "mariadb",
				"Config": {
					"db_password": "secret"
				}
			}
		],
		"Image": "ubuntu21.04",
		"NHosts": 1
	}'


Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса MariaDB доступна по API:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/mariadb


**PostgreSQl**

`PostgreSQL`_ это система управления реляционными базами данных (СУБД) с открытым исходным кодом, в которой особое внимание уделяется расширяемости и совместимости с SQL. Этот сервис можно было развернуть на узле хранения в облаке с помощью Michman. 

На текущий момент в системе поддерживаются версии *9.6*, *10*, *11* and *12* PostgreSQL. 

Параметр конфигурации сервиса **postgresql** включает:

* **db_password** -- Пароль пользователя по умолчанию для БД PostgreSQL для пользователя 'postgres', вы можете его изменить позже.

Следующий пример показывает запрос для развертывания в облаке СУБД PostgreSQl.

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "my-cluster",
		"Services": [
			{
				"Name": "postgresql-service",
				"Type": "postgresql",
				"Configs": {
					"db_password": "secret"
				}
			}
		],
		"Image": "ubuntu",
		"NHosts": 1
	}'

Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса PostgreSQL доступна по API:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/postgresql

**Redis**

`Redis`_ представляет собой хранилище структурных данных в памяти с открытым исходным кодом (под лицензией BSD), используемое в качестве базы данных, кеша и брокера сообщений. Этот сервис можно было развернуть на узле хранения в облаке с помощью Michman. 

На текущий момент поддерживается версия *latest* Redis. 

Параметр конфигурации сервиса **redis** включает:

* **db_password** -- Пароль пользователя по умолчанию для Redis, вы можете его изменить позже. Имя пользователя не является обязательным. 

Следующий пример показывает запрос для развертывания в облаке СУБД Redis.

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "my-cluster",
		"Services": [
			{
				"Name": "redis-service",
				"Type": "redis",
				"Configs": {
					"db_password": "secret"
				}
			}
		],
		"Image": "ubuntu",
		"NHosts": 1
	}'

Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса Redis доступна по API:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/redis

Обработка больших данных
-------------------------

Для сложных вычислительных задач и задач с обработкой больших данных могут использоваться такие приложения, как Apache Spark и Apache Hadoop, Slurm . 

**Apache Spark и Apache Hadoop** 

Apache `Spark`_ это единый аналитический инструмент для обработки больших данных со встроенными модулями для потоковой передачи, SQL, машинного обучения и обработки графиков.  Этот сервис может быть развернут в распределенном кластере виртуальных машин в облаке с помощью Michman.

Программная библиотека Apache `Hadoop`_ - это среда, которая реализует распределенную обработку больших объемов данных между кластерами компьютеров с использованием моделей программирования. Он предназначен для масштабирования от отдельных серверов до тысяч машин, каждая из которых предлагает локальные вычисления и хранение данных. 

Мичман запускает Spark, подключенный к Hadoop, и поддерживает различные плагины Spark: Jupyter, Jupyterhub, Cassandra. Также его можно запустить с помощью диспетчера ресурсов `YARN`_. 

На текущий момент поддерживаются следующие версии Spark: *1.0.0*, *1.0.1*, *1.0.2*, *1.1.0*, *1.1.1*, *1.2.0*, *1.2.1*, *1.2.2*, *1.3.0*, *1.3.1*, *1.4.0*, *1.4.1*, *1.5.0*, *1.5.1*, *1.5.2*, *1.6.0*, *1.6.1*, *1.6.2*, *2.0.0*, *2.0.1*, *2.0.2*, *2.1.0*, *2.2.0*, *2.2.1*, *2.3.0*.

Config parameter for **spark** service type supports:

* **use-yarn** -- режим развертывания Spark-on-YARN  (имеет накладные расходы на память, поэтому не используйте его, если не знаете зачем)
* **hadoop-version** -- выбор конкретной версии Hadoop для Spark. По умолчанию устанавливается последняя версия поддерживаемая Spark.
* **spark-worker-mem-mb** --  не следует определять автоматически рабочую память Spark и использовать указанное значение, может быть полезно, если другим процессам на  slave-узлах (например, python) требуется больше памяти, по умолчанию для slave-узлов ОЗУ 10–20 ГБ необходимо оставить 2 ГБ для системы/других процессов; 
* **yarn-master-mem-mb** -- объем физической памяти в MB, который может быть аллоцирован в контейнере. По умолчанию это значение 10240.
      
Следующий пример показывает запрос для создания кластера из трех узлов с сервисом  Apache Spark в режиме YARN.

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "my-cluster",
		"Services": [
			{
				"Name": "spark-service",
				"Type": "spark",
				"Version": "2.3.0",
				"Configs": {
					"use-yarn": "true"
				}
			}
		],
		"Image": "ubuntu",
		"NHosts": 3
	}'

Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса Apache Spark доступна по API:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/spark

**Slurm** 

`Slurm`_ - отказоустойчивая система распределения заданий и ресурсов на кластерах Linux и Unix-подобных ядер. При помощи системы оркестрации Michman диспетчер задач может быть развернут на виртуальном кластере. На данный момент развертывание Slurm доступно на базе операционной системы Ubuntu.

Поддерживается 2 версии сервиса. В зависимости от версии ОС устанвливается соответствующая версия Slurm:  если ОС  - ubuntu21.04, то Slurm version = slurm-wlm 20.11.4, соответственно  ubuntu18.04 - slurm-wlm 17.11.2. Slurm может быть развернут с системой логирования и без нее, с файловой системой NFS и без нее. Для этого в запросе на создание кластера необходимо указать соответствующую версию: Slurm - развертываемая версия по умолчанию без дополнительных сервисов, Slurm-db - Slurm будет развернут с системой логирования, Slurm-nfs - Slurm будет настроен совместно с файловой системой NFS, Slurm-db-nfs - Slurm c системой логирования и файловой системой NFS. Предоставляется REST API интерфейс для взаимодействия с Slurm-кластером.

Параметры, доступные пользователю для изменения конфигурации развертываемого сервиса Slurm: 

* **use_rest** -- пользователю предоставляется Slurm-кластер с REST API интерфейсом. Данный параметр может быть выставлен, если образ операционной системы - ubuntu21.04 и версия Slurm - Slurm-db. Значение по умолчанию: false. Для корректной работы Slurm REST API пользователь должен экспортировать переменную оболочки SLURM_JWT с заранее сгенерированным значением на тот хост, с которого будет отправлен запрос. Для этого необходимо зайти на master-хост, скопировать содержимое файла /var/log/slurm/slurm_token в командную строку (выполнить SLURM_JWT= ...). В запросе к Slurm REST API надо указать переменные X-SLURM-USER-NAME и X-SLURM-USER-TOKEN, значения которых строго фиксированы: X-SLURM-USER-NAME:root и X-SLURM-USER-TOKEN:${SLURM_JWT}. 

	Пример запроса: 
	
	.. parsed-literal::
		curl -H "X-SLURM-USER-NAME:root" -H "X-SLURM-USER-TOKEN:${SLURM_JWT}" http://{IP-адрес master-хоста}:6820/slurm/v0.0.36/ping
	
	Примеры запросов предствлены здесь: https://app.swaggerhub.com/apis/rherrick/slurm-rest_api/0.0.35.

* **db_password** -- пароль создаваемой базы данных для системы логирования. Данный параметр доступен пользоателю при указании версии Slurm-db. Значение по умолчанию: slurmdbd
* **db_user** -- пользователь создаваемой базы данных системы логирования. Данный параметр доступен пользоателю при указании версии Slurm-db. Значение по умолчанию: slurm
* **TaskPluginParam** -- параметр конфигурационного файла slurm.conf. Параметр для TaskPlugin, который определяет тип подключаемого модуля запуска задач, используемого для управления ресурсами в узле. Допустимые значения: None, Boards, Sockets, Cores, Threads, и/или Verbose. При указании нескольких значений, они должны быть разделены запятой. Значение по умолчанию: None.
* **use_open_foam** -- пользователю предоставляется Slurm-кластер с установленным на всех хоcтах OpenFOAM.
* **config_dir** -- путь к шаблону конфигурационного файла slurm.conf.
* **cgroup_config_dir** -- путь к шаблону конфигурационног файла cgroup.conf
* **use_open_mpi** -- пользователю предоставляется Slurm-кластер с установленной на master-хосте и slave-хостах библиотекой OpenMPI.
* **partitions** -- параметр, описывающий разделение Slurm-кластера. Данные конфигурации находятся в файле slurm.conf. Список состоит из строк, аргументы которых разделены символом ':'. Первый аргумент - имя раздела, второй - количество хостов, относящихся к этому разделу. Раздел с названием "main" должен бфть в каждом пользовательском запросе, так как используется как дефолтный. Пример списка из пользовательского запроса: \"main:5\", \"part_1:2\", \"part_2:3\", \"part_3:4\"  
* **open_mpi_version** -- версия устанавливаемой библиотеки OpenMPI.
Следующий пример показывает запрос для создания Slurm-кластера из двух узлов с системой логирования и интерфейсом REST API: 

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
 	'{
		"DisplayName":"test", 
		"Services":[{
			"Name":"Slurm service",
			"Type":"slurm",
			"Version": "Slurm-db",
			"Config":{
				"use_rest": "true"
			}
		}], 
		"Description": "cluster", 
		"Image": "ubuntu21.04", 
		"NHosts": 2
	}'

Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса Slurm доступна по API:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/slurm

Web-лаборатории
----------------

С Michman легко можно развернуть самые популярные веб-лаборатории для интерактивной разработки: Jupyter и Jupyterhub. 

**Jupyter**

`JupyterLab`_ это интерактивная веб-среда разработки для ноутбуков Jupyter, кода и данных. JupyterLab отличается гибкостью: он имеет настраиваемый пользовательский интерфейс и может быть использован в широком спектре рабочих процессов в области науки о данных, научных вычислений и машинного обучения. Его можно развернуть на master-узле в облаке при помощи Michman. 

На текущий момент поддерживается версия *6.0.1*. Он также может быть развернут вместе с плагином Spark-connector.

Следующий пример показывает запрос для создания кластера из трех узлов с сервисом Jupyter.

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "my-cluster",
		"Services": [
			{
				"Name": "jupyter-service",
				"Type": "jupyter"
			}
		],
		"Image": "ubuntu",
		"NHosts": 3
	}'

Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса Jupyter доступна по API:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/jupyter

**JupyterHub**

`JupyterHub`_ предоставляет возможности ноутбуков Jupyter для группового использования. Он предоставляет пользователям доступ к вычислительным средам и ресурсам. Пользователи, в том числе студенты, исследователи и специалисты по данным, могут выполнять свою работу в своих собственных рабочих областях на общих ресурсах, которыми могут эффективно управлять системные администраторы.

На текущий момент поддерживается версия *1.3.0* Jupyterhub. Он также может быть развернут вместе с плагином Spark-connector.

Следующий пример показывает запрос для создания кластера из трех узлов с сервисом JupyterHub.

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "my-cluster",
		"Services": [
			{
				"Name": "jupyterhub-service",
				"Type": "jupyterhub"
			}
		],
		"Image": "ubuntu",
		"NHosts": 3
	}'

Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса JupyterHub доступна по API:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/jupyterhub


Управление файлами
--------------------

Michman предоставляет сервисы для удобной работы с данными в облаке. 

**NextCloud**

`Nextcloud`_ представляет собой набор клиент-серверного программного обеспечения для создания и использования услуг хостинга файлов. Nextcloud является системой с открытым исходным кодом, что означает, что любой может установить и использовать его на своих частных серверных устройствах. Его можно развернуть на узле хранения в облаке с помощью Michman.

Параметр конфигурации для типа сервиса **nextcloud** включает следующие поля:

* **weblab_name** -- имя Web-лаборатории.
* **nfs_server_ip** -- IP NFS-сервера.
* **mariadb_image** -- образ docker для mariadb
* **nextcloud_image** -- образ docker для nextcloud


Следующий пример показывает запрос для создания кластера из трех узлов с сервисом Nextcloud.

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "my-cluster",
		"Services": [
			{
				"Name": "nextcloud-service",
				"Type": "nextcloud"
			}
		],
		"Image": "ubuntu",
		"NHosts": 3
	}'

Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса Nextcloud доступна по API:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/nextcloud

**NFS Server**

NFS (Network File Share) это протокол, который позволяет обмениваться каталогами и файлами с другими клиентами Linux в сети. Каталог для совместного использования обычно создается на сервере NFS, и файлы добавляются в него. 

Клиентские системы монтируют каталог, находящийся на сервере NFS, который предоставляет им доступ к созданным файлам. NFS пригодится, если нужно поделиться общими данными между клиентскими системами, особенно когда им не хватает места. 

Параметр конфигурации для типа сервиса **nfs-server** включает следующее поле:
* **weblab_name** -- имя Web-Лаборатории.

Следующий пример показывает запрос для создания кластера из трех узлов с сервисом NFS Server.

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "my-cluster",
		"Services": [
			{
				"Name": "nfs-server",
				"Type": "nfs"
			}
		],
		"Image": "ubuntu",
		"NHosts": 3
	}'


Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса NFS Server доступна по API:

.. parsed-literal::
	curl http://michman_addr:michman_port/configs/nfs

Управление логами
-----------------

Michman позволяет развернуть в облаке стандартный стек технологий для обработки и хранения логов.

**Elasticsearch**

`Elasticsearch`_ это поисковая система, основанная на библиотеке Lucene. Он предоставляет распределенную, многопользовательскую полнотекстовую поисковую систему с веб-интерфейсом HTTP и документами JSON без определенных схем. Его можно развернуть в распределенном кластере виртуальных машин в облаке с помощью Michman. 

На текущий момент поддерживается версия *7.1.1* Elasticsearch.

Параметр конфигурации для типа сервиса **elastic** включает следующее поле:
* **heap-size** -- настраивает определенный размер кучи ElasticSearch в ГБ. Размер кучи по умолчанию - 1 ГБ. 

Следующий пример показывает запрос для создания кластера из трех узлов с сервисом Elasticsearch.

.. parsed-literal::
	curl http://michman_addr:michman_port/projects/{ProjectID}/clusters -XPOST -d 
	'{
		"DisplayName": "my-cluster",
		"Services": [
			{
				"Name": "elastic-server",
				"Type": "elastic"
			}
		],
		"Image": "ubuntu",
		"NHosts": 3
	}'

Полная и актуальная информация о поддерживаемых параметрах конфигурации и версиях сервиса Elasticsearch доступна по API:


.. parsed-literal::
	curl http://michman_addr:michman_port/configs/elastic

Ближайшие планы
----------------

В 2021 планируется добавление поддержи в Michman следующих сервисов, которые могут быть развернуты в облаке:

* `Kubernetes`_ -- система контейнерной оркестрации для автоматизации развертывания вычислительных приложений, масштабирования и управления. 