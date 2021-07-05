.. _michman_configuration_section:

.. _HYDRA: https://www.ory.sh/hydra/docs/
.. _Werther: https://github.com/i-core/werther
.. _README: https://github.com/i-core/werther/blob/master/README.md
.. _спецификацией: https://tools.ietf.org/html/rfc6749#section-4.1.3


Начальная конфигурация Michman
===============================

В этом разделе представлена информация о начальной конфигурации Michman и описаны подготовительные этапы для начала работы с системой. Michman зависит от различных компонентов, которые отвечают за определенные функции, такие как хранение данных, аутентификация, ведение журнала и т. д. Michman должен быть подключен к облаку IaaS. На текущий момент Michman поддерживает развертывание кластеров в облаках на базе OpenStack. 

Настройка информации об облаке на базе OpenStack 
-------------------------------------------------

Мичман использует служебный аккаунт в облаке OpenStack, все кластеры развертываются в одном проекте, указанном пользователем. При этом для каждого кластера создается отдельная группа безопасности. 

Сейчас поддерживаются следующие версии OpenStack:
	* **Stein**
	* **Liberty**
	* **Ussuri**

В настоящее время оркестратор поддерживает развертывание сервисов только на виртуальных машинах с ОС Ubuntu (16.04 или 18.04) или CentOS, поэтому необходимо заранее подготовить в облаке подходящие образы. 

Рекомендуется также подготовить пул плавающих IP-адресов и типы экземпляров ВМ для будущих виртуальных машин. 

Также нужно подготовить пару ключей безопасности и pem-ключ для обеспечения доступа к созданным виртуальным машинам из сервиса Michman *launcher*. Ключ должен быть добавлен в файл `$PROJECT_ROOT/launcher/ansible/files/ssh_key` или в хранилище секретов Vault.

Далее необходимо загрузить файл OpenStack RC и записать информацию о доступе к облаку, которая в нем указана, в хранилище секретов Vault. Конкретные поля для каждой версии перечислены ниже. 

Укажите следующие параметры облака в файле *config.yaml*:
	
	* `os\_key\_name: OS\_KEY\_NAME`
	* `virtual\_network: NETWORK`
	* `floating\_ip\_pool: IP\_POOL`
	* `master\_flavor: FLAVOR`
	* `slaves\_flavor: FLAVOR`
	* `storage\_flavor: FLAVOR`
	* `os\_version: VERSION #stein or liberty or ussuri`


Конфигурация хранилища секретов
-------------------------------
В Michman используется система хранения секретов Vault для безопасного доступа к конфиденциальным данным, таким как учетные данные базы данных, аутентификационные данные облака и т. д. 

Используемая версия: 1.2.3

Необходимо записать в Vault следующие секреты (Тип для *secret engine* - kv v1, путь: kv/).
    
Секрет Openstack (os_key) включает следующие ключи для версии **Liberty**:
	* **OS_AUTH_URL**
	* **OS_PASSWORD**
	* **OS_PROJECT_NAME**
	* **OS_REGION_NAME**
	* **OS_TENANT_ID**
	* **OS_TENANT_NAME**
	* **OS_USERNAME** 
	* **OS_SWIFT_USERNAME** -- optional
	* **OS_SWIFT_PASSWORD** -- optional 

Секрет Openstack (os_key) включает следующие ключи для версии **Stein**:
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

Секрет Openstack (os_key) включает следующие ключи для версии **Ussuri**:
	* **OS_AUTH_URL**
	* **OS_PASSWORD**
	* **OS_PROJECT_NAME**
	* **OS_PROJECT_ID**
	* **OS_REGION_NAME**
	* **OS_USERNAME** 
	* **OS_IDENTITY_API_VERSION**
	* **OS_PROJECT_DOMAIN_ID**
	* **OS_USER_DOMAIN_NAME**
	* **OS_INTERFACE**


Секрет с ключом Ssh (ssh_key) содержит слежующий ключ:
	* **key_bgt** -- private ssh key for Ansible commands

Секрет Couchbase (cb_key) включает следующие ключи:
	* **password** -- пароль от couchbase
	* **path** -- адрес couchbase
	* **username** -- имя пользователя couchbase 

Секрет репозитория Docker (registry_key) включает следующие ключи:
	* **url** -- адрес вашего самоподписанного репозитория в соответствии с сертификатом или URL-адрес репозитория gitlab 
	* **user** -- ваш самоподписанный репозиторий или имя пользователя gitlab
	* **password** -- ваш самоподписанный репозиторий или пароль gitlab

Этот секрет является опциональным.

Секрет Hydra (hydra_key) включают следующие ключи:
	* **redirect_uri** -- OAuth 2.0 redirect URI
	* **client_id** -- OAuth 2.0 client ID
	* **client_secret** -- OAuth 2.0 client secret

Этот секрет является опциональным и используется только если выбрана модель авторизации oauth2.

Также необходимо указать следующие параметры Vault в файле *config.yaml*:

	* `token: ROOT\_TOKEN`
	* `vault\_addr: VAULT\_ADDR`
	* `os\_key: BUCKET\_PATH`
	* `cb\_key: BUCKET\_PATH`
	* `ssh\_key: BUCKET\_PATH`
	* `hydra\_key: BUCKET\_PATH`

Конфигурация базы данных
-------------------------

Для хранения данных о созданных в системе кластерах, проектах, шаблонах, а также о доступных для развертывания сервисов и образов ОС в Michman используется Couchbase Server.

Используемая версия: 6.0.0 community edition.

Необходимо создать следующие бакеты с первичными индексами: **clusters**, **projects**, **templates**, **service_types**, **images**. 
развенуть
Для корректной работы оркестратора перед началом работы с Michman и созданием кластеров необходимо заполнить бакеты **service_types** и **images**. Для этого рекомендуется использовать *REST API*.

Зарегистрируйте сервисы, которые вы планируете развертывать при помощи Michman. Описания доступных на текущий момент сервисов в формате Json перечислены в директории *init*. Например, для регистрации типа сервиса *spark* необходимо выполнить следующий запрос:

.. parsed-literal::
	curl -X POST -d "data=@michman/init/spark.json" http://michman_addr:michman_port/configs


Зарегистрируйте облачные образы ОС, которые вы планируете использовать в кластерах. Эти образы должны быть предварительно созданы в облаке OpenStack. Например, для регистрации образа *ubuntu* необходимо выполнить следующий запрос: 

.. parsed-literal::
	curl -X POST http://michman_addr:michman_port/configs -d 
	'{
		"Name": "ubuntu",
		"AnsibleUser": "ubuntu",
		"CloudImageID": "UUID"
	}'

Также перед началом работы с Michman можно создать пользовательские проекты и общесистемные шаблоны кластеров при помощи REST API.

Конфигурация логирования
-------------------------

Michman производит три типа логов: логи rest-сервисв, логи launcher-сервиса и логи процесса развертывания кластера.


Логи сервисов rest и launcher хранятся в файлах в директории `$PROJECT_ROOT/logs` и доступны по REST API.

Логи развертывания кластера это логи, которые производятся системой Ansible на процессах создания, обновления и удаления кластера. Логи кластера могут храниться в директории, указанной пользователем, или в сервисе Logstash.

Для хранения логов кластера в файлах, укажите следующие поля в *config.yaml*:

	* `logs\_output: file`
	* `log_file\_path: PATH`

Для хранения логов кластера в хранилище Logstash, необходимо развернуть сервисы Logstash и Elasticsearch. Опционально может быть развернута система Kibana.

Обновите конфигурационный файл Logstash config.conf: 

.. parsed-literal::
	input{
		http {
	    		host => "0.0.0.0" 
	    		port => 9000
	  	}
	}
	filter{
		mutate { 
			add_field => { "[@metadata][target_index]" => "%{Cluster_name}" } 
			remove_field => [ "Cluster_name" ] 
		}
	}
	output {
		elasticsearch {
				hosts => ["<ELASTICSEARCH\_ADDR>:9200"]
				index => "%{[@metadata][target_index]}"
		}
	} 

Далее укажите адресы Logstash и Elasticsearch в config.yaml файле Michman:

	* `logs\_output: logstash`
	* `logstash\_addr: xx.xx.xx.xx:xxxx`
	* `elastic\_addr: xx.xx.xx.xx:xxxx`

Логи кластера далее могут быть доступны по REST API по ID кластера.

Конфигурация Docker репозитория
-------------------------------

Текущее развертывание сервиса Nextcloud основано на контейнерах Docker. В случае использования локального репозитория Docker необходимо выполнить следующие шаги.


    #. Подготовьте репозиторий. Это может быть небезопасный репозиторий (без каких-либо сертификатов и пользовательских элементов управления), самоподписанный репозиторий или репозиторий gitlab. 
    #. Укажите в *config.yaml*:

    	#. Для небезопасного репозитория заполните следующие поля:

    		* `docker\_incecure\_registry: true`
    		* `insecure\_registry\_ip: xx.xx.xx.xx:xxxx`

    	#. Для самоподписанного репозитория укажите следующие значения:

    		* `docker\_selfsigned\_registry: true`
    		* `docker\_selfsigned\_registry\_ip: xx.xx.xx.xx:xxxx`
    		* `docker\_selfsigned\_registry\_url: consides.to.cert.url`
    		* `docker_cert_path: path_to_registry_cert.crt`

      	#. В случае использования репозитория gitlab, укажите:

      		* `docker\_gitlab\_registry: true`

    #. В случае, если используется самоподписанный репозиторий или gitlab репозиторий, в **Vault** необходимо указать секрет с ключами *url*, *user* и *password* и указать в *config.yaml*:

    	* `registry\_key: key\_of\_docker\_secret` 

Настройки аутентификации и авторизации
------------------------------------------

Внутренняя модель представлений данных Michman подразумевает логическое разделение кластеров на группы внутри проектов системы. Пользователи могут получить доступ к информации о кластерах только из тех проектов, членами которых они являются. На основе такого разделения в Michman реализованы три роли:

	* **admin** - администратор Michman, может создавать новые проекты, добавлять информацию о доступных для развертывания сервисов, добавлять общедоступные шаблоны кластеров. 
	* **user** - анонимный пользователь, имеет доступ на чтение для путей, не связанных с конкретными проектами Michman.
	* **project_member** - член проекта, может создавать новые кластеры, изменять, удалять их и получать информацию о кластерах, в рамках своего проекта. 

Michman не хранит информацию о пользователях и их группах, аутентификация осуществляется при помощи сторонних сервисов. На текущий момент поддерживаются следующие модели аутентификации:

	* **OAUTH2.0**
	* **OpenStack Keystone**
	* **None-authentication mode**

В следующих секциях подробно рассматривается каждая из этих моделей.

**Аутентификация OAUTH2.0**

Поток аутентификации OAuth2.0 реализован в Michman при помощи следующих систем: 

	* ORY `HYDRA`_ -- реализация фреймворка авторизации OAuth 2.0 фреймворка OpenID Connect Core 1.0.
	* `Werther`_ -- Identity Provider для ORY Hydra поверх LDAP. Реализует потоки Login и Consent и предоставляет базовый UI.

Этот тип аутентификации используется для того, чтобы использовать Michman с LDAP-сервером - пользователи получают доступ к Michman со своими логинами LDAP, а информация о группах пользователей извлекается из групп LDAP. 

Необходимо развернуть следующие сервисы: Hydra Admin, Hydra Client и Werther, взаимодействующий с вашим LDAP. Самый простой способ развернуть эти системы -- воспользоваться файлом docker-compose, описанным в Wearther `README`_.

Замечание! Необходимо настроить следующие параметры окружения Werther:

	* **WERTHER_LDAP_ROLE_CLAIM**
	* **WERTHER_IDENTP_CLAIM_SCOPES**
	* **WERTHER_LDAP_ATTR_CLAIMS**

Обязательно следует указать параметр "groups", который будет использоваться для авторизации пользователя в проектах Michman.

Замечание! Необходимо настроить следующие параметры окружения Hydra Admin:

	* **WEBFINGER_OIDC_DISCOVERY_SUPPORTED_SCOPES**
	* **WEBFINGER_OIDC_DISCOVERY_SUPPORTED_CLAIMS**

Обязательно следует указать параметр "groups" в *scopes* и *claims* Oauth2, который будет использоваться для авторизации пользователя в проектах Michman.

Замечание! При запуске команды "hydra clients create" необходимо указать следующие параметры:

    * grant\-types
    * token\-endpoint\-auth\-method
    * scope 
    * callbacks 
    * post\-logout\-callbacks
    * response\-types

Команда запуска должна быть похожа на следующую:

.. parsed-literal::

	hydra clients create \
	 --skip-tls-verify \
     --id test-client \
     --secret test-secret \
     --response-types code,id_token \
     --grant-types authorization_code \
     --token-endpoint-auth-method client_secret_post \
     --scope openid,profile,email,groups \
     --callbacks http://michman_addr:michman_port/auth \
     --post-logout-callbacks http://michman_addr:michman_port/auth

Обязательно следует указать параметр "groups" в *scopes* Oauth2.

После развертывания указанных сервисов может быть пройдена аутентификация и авторизация. Этот процесс состоит из следующих шагов:

	#. Отправьте запрос аутентификации в сервис Hydra Client с кодом grant_type, параметр *groups* должен быть указан в Oauth2 scopes. Также в scopes должен быть указан параметр openid, остальные поля необязательны (указаны здесь в качестве примера): 

	.. parsed-literal::
		
		http://hydra_client:4444/oauth2/auth?client_id=test-client&response_type=code&scope=openid%20profile%20email%20groups&state=12345678

	#. По запросу вы будете перенаправлены на форму входа Werther в браузере. Нужно ввести логин-пароль пользователя от учетной записи в LDAP. В случае успеха он перенаправляется на путь */auth* в Michman. В параметры запроса будет добавлен код аутентификации. 

	#. Продолжение аутентификации и авторизации обрабатывается в Michman: 
		
		#. Параметр “code” извлекается из параметров запроса. 
		#. Формируется POST-запрос для получения токена на адрес hydra-client:4444/auth2/token, в соответствии со `спецификацией`_.
		#. Обработанный ответ в случае успеха содержит токен доступа в теле ответа.
		#. Также формируется GET-запрос на адрес hydra-client:4444/userinfo. Устанавливается заголовок авторизации, который содержит полученный ранее токен. В случае успеха: 

			* информация о группах пользователя извлекается из ответа на запрос *userinfo*;
			* для пользователя устанавливается новая сессия;
			* группы пользователя и токен доступа сохраняется в параметры сессии. 

После этого процесса вы сможете получить доступ к проектам, связанным с вашими группами, и создавать в них новые кластеры. Если группа "admin" присутствует в списке групп, вы можете получить доступ к действиям администратора.

Без аутентификации вы получите роль «user». 

Также необходимо заполнить следующие поля в файле *config.yaml*:

	* `use\_auth: true`
	* `authorization\_model: oauth2`
	* `admin\_group: admin`
	* `session\_idle\_timeout: 480 #time in minutes, controls the maximum length of time a session can be inactive before it expires`
	* `session\_lifetime: 960 #time in minutes, controls the maximum length of time that a session is valid for before it expires`

	* `hydra\_admin: HYDRA\_ADDR`
	* `hydra\_client: HYDRA\_ADDR`

**Аутентификация Keystone**

Для этого типа аутентификации необходимо иметь аккаунт в системе OpenStack Keystone. Пройдите аутентификацию в Keystone и получите следующие токены:

	* **X-Auth-Token**
	* **X-Subject-Token**

Далее можно начать процесс аутентификации и авторизации:
	
	#. Перейдите на http://michman_addr:michman_port/auth, указав токены *X-Auth-Token* и *X-Subject-Token* в заголовках.
	#. Оставшийся процесс обрабатывается в Michman. В нем отправляется запрос к сервису Keystone на адрес: `keystone\_addr:keystone\_port/v3/auth/tokens` и извлекается информация из ответа о ролях пользователя. Роли пользователя будут сохранены в параметр *groups* в сессии пользователя. 

После этого процесса вы сможете получить доступ к проектам, связанным с вашими группами, и создавать в них новые кластеры. Если группа "admin" присутствует в списке групп, вы можете получить доступ к действиям администратора.

Без аутентификации вы получите роль «user». 

Также необходимо заполнить следующие поля в файле *config.yaml*:

	* `use\_auth: true`
	* `authorization\_model: keystone`
	* `admin\_group: admin`
	* `session\_idle\_timeout: 480 #time in minutes, controls the maximum length of time a session can be inactive before it expires`
	* `session\_lifetime: 960 #time in minutes, controls the maximum length of time that a session is valid for before it expires`
	* `keystone\_addr: KEYSTONE\_ADDR`

**Режим администратора**

In addition, Michman supports none authentication mode, which could be used, for example, for development purposes. In this mode every user after authentication obtains "admin" role.
Кроме того, Michman поддерживает режим аутентификации, в котором все пользователи получают роль администратора. Такой режим может использоваться, например, в целях разработки.

Он включает следующие шаги:

	#. Перейдите по адресу http://michman_addr:michman_port/auth.
	#. Остающийся процесс обрабатывается в Michman. Он устанавливает новый сеанс пользователя и сохраняет группу «admin» в параметре groups в сессии пользователя. 

Также необходимо заполнить следующие поля в файле *config.yaml*:

	* `use\_auth: true`
	* `authorization\_model: none`
	* `admin\_group: admin`
	* `session\_idle\_timeout: 480 #time in minutes, controls the maximum length of time a session can be inactive before it expires`
	* `session\_lifetime: 960 #time in minutes, controls the maximum length of time that a session is valid for before it expires`

**Отключение аутентификации и авторизации**

Вы можете полностью отключить аутентификацию и авторизацию в системе Michman и работать с Michman без установления сессии. 

Также необходимо заполнить следующее поле в файле *config.yaml*:

	* `use\_auth: false`