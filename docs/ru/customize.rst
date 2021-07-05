.. _michman_customize_section:

.. _проекта: https://github.com/ispras/spark-openstack

Модификация поддерживаемых сервисов
===================================

В этой секции рассказывается, как изменять список поддерживаемых в Michman сервисов: добавлять новые версии, конфигурации для них и другое. Также здесь описывается процесс добавления новых поддерживаемых сервисов в Michman. Этот документ будет полезен для администраторов и разработчиков Michman.

This section сovers how to customize Michmans services: add new versions, configuartion options and more. It also describes how to add new services to Michman. This guide will be useful to Michman's administrators and developers.


Описание основного Ansible-playbook 
-----------------------------------

Ansible-playbook, используемый для создания виртуальной инфраструктуры и развертывания сервисов, является расширением и развитием `проекта`_ Spark-Openstack.

Для операций "create", "update" и "delete" запускается плейбук *main.yml*. В общем, выполнение сценария состоит из следующих шагов:

	#. Создание кластера. A
	Creation of the cluster. Ansible запускает роль для запуска виртуальных машин в облаке на основе запрошенных конфигураций кластера, создает группу безопасности и предоставляет плавающие IP-адреса созданным виртуальным машинам. По умолчанию для создания виртуальных машин используется асинхронный режим. Затем ssh-ключ пользователя добавляется на созданные виртуальные машины. Этот шаг выполняется при операции "create". 

	#. На втором этапе развертываются некоторые базовые пакеты и конфигурации, перечисленные в роли *base*.

	#. Затем развертываются сервисы. Сервисы равзертываются на разных группах хостов. На данный момент поддерживаются три группы:
		* master
		* slave
		* storage

     Ansible включает одну или несколько ролей Ansible для каждого запрошенного сервиса. Если переменные, используемые в роли сервиса, настраиваются через Michman API, их имя должно быть соответствовать следующему шаблону:
	 .. parsed-literal::
	 	<service_name>_<parameter_name>

	 Если для сервиса поддерживается выбор различных версий, то для версии должна быть определена такая переменная:

	 .. parsed-literal::
	 	<service_name>_version

	 Если у роли есть зависимости, то они должны быть записаны в разделе "dependencies" в *meta*:

	 .. parsed-literal::
		---
		dependencies:
		 	- { role: spark_common }

Конфигурация поддерживаемых сервисов
------------------------------------

Вы можете изменить или добавить следующую информацию в поддерживаемые в Michman сервисы (*service type*):
	* DefaultVersion -- версия сервиса по умолчанию
	* Description -- описание сервиса
	* AccessPort -- порт доступа к сервису по умолчанию, который будет указан в URL доступа к сервису
	* Ports -- можно добавлять информацию о других портах доступа для сервиса
	* Supported versions -- список подедрживаемых сервисом версий

Для этого используется метод Michman PUT */configs/{serviceType}* и приводится JSON с изменениями. Например:

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

Вы можете изменить или добавить следующую информацию для версии сервиса:
	* Description -- описание версии сервиса
	* Configuration parameters -- настраиваемые переменные для версии сервиса
	* Описание зависимостей для этой версии от других версий других поддерживаемых сервисов

Для этого используется метод Michman PUT */configs/{serviceType}/versions/{versionId}* и приводится JSON с изменениями. Например:

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
	    
.. note:: Также необходимо добавить возможность развертывания новых версий сервиса или настройки переменных в роли Ansible для этого сервиса.

Добавление нового сервиса
--------------------------

В этом разделе описывается добавление нового сервиса в Michman на примере СУБД Apache Ignite. Регистрация сервиса включает следующие шаги.

 	#. **Добавление роли Ansible для развертывания Apache Ignite.** Имя роли должно соответствовать типу зарегистрированного сервиса (в данном случае ignite), все настраиваемые пользователем переменные для этой роли должны иметь в своем имени префикс с типом сервиса.

 	#. **Описание зарегистрированного типа сервиса в формате JSON.** Сервис должен описывать информацию о поддерживаемых версиях, настраиваемых параметрах и зависимостях. Также указывается класс сервиса и возможность доступа к нему. Ниже приведен пример JSON, описывающей сервис Apache Ignite с поддерживаемой версией 7.1.1 и настраиваемым размером рабочей памяти. Поле Class описывает связь между сервисом и инфраструктурой. В этом примере master-slave означает, что сервис развернут в распределенном режиме. Для вашего удобства мы рекомендуем добавить этот документ JSON в директорию *init*.

	#. **Запрос на регистрацию нового поддерживаемого сервиса.** Администратор Michman должен отправить следующий запрос:

.. parsed-literal::
	curl -X POST -d @michman/init/ignite.json http://michman_addr:michman_port/configs

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

В случае успешной регистрации новогоы сервиса пользователю возвращается ответ, содержащий HTTP-код 200 и JSON с дополненным описанием типа сервиса «ignite».
