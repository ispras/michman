.. _michman_install_section:

.. _ссылке: https://medium.com/better-programming/install-go-1-11-on-ubuntu-18-04-16-04-lts-8c098c503c5f

.. _здесь: https://askubuntu.com/questions/1072683/how-can-i-install-protoc-on-ubuntu-16-04

.. _тут: https://hub.docker.com/r/znly/protoc/

Установка Michman
==================

В этом разделе приводится инструкция по установке Michman.

Установка Michman на операционной системе Ubuntu 
------------------------------------------------
Необходимо установить следующие apt-пакеты:

.. parsed-literal::

    sudo add-apt-repository ppa:longsleep/golang-backports
	sudo apt update
	sudo apt install golang-go
	sudo apt install unzip apt-transport-https
	sudo apt install ca-certificates curl software-properties-common
	sudo apt install python python-pip python-setuptools

Также необходимо установить:
	* python and python-pip (or python3/python3-pip)
	* zip/unzip
	* openssh-client
	* wget 

.. note:: Перед установкой Michman необходимо установить окружение для ЯП Go и выполнить конфигурацию рабочей директории для проектов Go. Пример установки Go и настройки окружения доступен по `ссылке`_. 

Потребуется установить следующие python-пакеты:

.. parsed-literal::
	pip install ansible==2.9.4 openstacksdk==0.40.0 # latest tested versions or
	# pip3 install ansible==2.9.4 openstacksdk==0.40.0

Также необходимо установить пакеты Go:

.. parsed-literal::
	go get -u google.golang.org/grpc
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/hashicorp/vault/api
	go get -u gopkg.in/yaml.v2
	go get github.com/julienschmidt/httprouter
	go get -u gopkg.in/couchbase/gocb.v1
	go get github.com/google/uuid
	go get github.com/golang/mock/gomock
	go get github.com/golang/mock/mockgen
	go get github.com/alexedwards/scs
	go get github.com/casbin/casbin

Скопируйте код проекта из github и переместите его в $GOPATH:

.. parsed-literal::
	git clone https://github.com/ispras/michman.git
	mkdir $GOPATH/src/github.com
	mkdir $GOPATH/src/github.com/ispras
	mv ./michman $GOPATH/src/github.com/ispras/
	cd $GOPATH/src/github.com/ispras/michman

Кроме того, необходимо установить библиотеку `libprotoc 3.6.1`. Инструкцию по установке можно найти `здесь`_ или может быть использован контейнер docker, как описано `тут`_.


Пример использования `libprotoc 3.6.1` при помощи docker-контейнера:

.. parsed-literal::
	docker pull znly/protoc
	cd $GOPATH/src/github.com/ispras/michman/protobuf
	docker run --rm -v $(pwd):$(pwd) -w $(pwd) znly/protoc --go_out=plugins=grpc:. -I. protofile.proto



Установка Michman в контейнерах
--------------------------------

Репозиторий Michman содержит файлы Dockerfile для сервисов *rest* и *launcher* для ОС **Ubuntu18.04** и **Centos8**, которые могут быть собраны и запущены в контейнерах **Docker** или **Podman**. 

Для начала необходимо скопировать репозиторий Michman:

.. parsed-literal::
	git clone https://github.com/ispras/michman.git
	cd michman

Затем необходимо заполнить конфигурационный файл Michman *config.yaml* с используемыми настройками.

.. note:: О том, как заполняется конфигурационный файл Michman, можно прочитать в следующей секции.

Следующие инструкции приводятся для сборки Michman поверх образа Ubuntu18.04. Для того, чтобы собрать Michman поверх Centos8, необходимо заменить путь **Ubuntu18.04** на **Centos8**.

**Инструкция для Docker**

	1. Соберите и запустите сервис michman-rest.

	.. parsed-literal::
		docker build -t michman-rest -f ./rest/Ubuntu18.04/Dockerfile .
		docker run --env CONFIG=/config.yaml --env LAUNCHER=localhost:5000 --env PORT=8081 -p 8081:8081 -v <path_to_config>/config.yaml:/config.yaml michman-rest

	2. Соберите и запустите сервис michman-launcher.

	.. parsed-literal::
		docker build -t michman-launcher -f ./launcher/Ubuntu18.04/Dockerfile .
		docker run -p 5000:5000 --env CONFIG=/config.yaml --env PORT=5000 -v <path_to_config>/config.yaml:/config.yaml michman-launcher

**Инструкция для Podman**

	
	1. Соберите и запустите сервис michman-rest.

	.. parsed-literal::
		podman build -t michman-rest -f ./rest/Ubuntu18.04/Dockerfile .
		podman run --env CONFIG=/config.yaml --env LAUNCHER=localhost:5000 --env PORT=8081 -p 8081:8081 -v <path_to_config>/config.yaml:/config.yaml:z michman-rest

	2. Соберите и запустите сервис michman-launcher.

	.. parsed-literal::
		podman build -t michman-launcher -f ./launcher/Ubuntu18.04/Dockerfile .
		podman run -p 5000:5000 --env CONFIG=/config.yaml --env PORT=5000 -v <path_to_config>/config.yaml:/config.yaml:z michman-launcher



