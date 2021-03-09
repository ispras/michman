.. _michman_install_section:

.. _link: https://medium.com/better-programming/install-go-1-11-on-ubuntu-18-04-16-04-lts-8c098c503c5f

.. _here: https://askubuntu.com/questions/1072683/how-can-i-install-protoc-on-ubuntu-16-04

.. _this: https://hub.docker.com/r/znly/protoc/

Installation Guide
===================

Welcome to the Michman Installation Guide!

Install Michman on Ubuntu 
--------------------------
Following apt-packages are required to perform the installation:

.. parsed-literal::

    sudo add-apt-repository ppa:longsleep/golang-backports
	sudo apt update
	sudo apt install golang-go
	sudo apt install unzip apt-transport-https
	sudo apt install ca-certificates curl software-properties-common
	sudo apt install python python-pip python-setuptools

Also are required:
	* python and python-pip (or python3/python3-pip)
	* zip/unzip
	* openssh-client
	* wget 

.. note:: You have to set up Go environment and configure your work directory for Go-projects. See this `link`_. 

Install following python-packages:

.. parsed-literal::
	pip install ansible==2.9.4 openstacksdk==0.40.0 # latest tested versions or
	# pip3 install ansible==2.9.4 openstacksdk==0.40.0

Go packages:

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

Clone project code from github and place it in $GOPATH:

.. parsed-literal::
	git clone https://github.com/ispras/michman.git
	mkdir $GOPATH/src/github.com
	mkdir $GOPATH/src/github.com/ispras
	mv ./michman $GOPATH/src/github.com/ispras/
	cd $GOPATH/src/github.com/ispras/michman

Also, `libprotoc 3.6.1` is required. Working installation described `here`_ or may be used docker container like `this`_.


Example:

.. parsed-literal::
	docker pull znly/protoc
	cd $GOPATH/src/github.com/ispras/michman/protobuf
	docker run --rm -v $(pwd):$(pwd) -w $(pwd) znly/protoc --go_out=plugins=grpc:. -I. protofile.proto



Install Michman in containers
------------------------------

We provide Dockerfiles for Michmans *rest* and *launcher* services for **Ubuntu18.04** and **Centos8**, which could be launched with **Docker** or **Podman** containers. 

First, you have to clone Michman repo:

.. parsed-literal::
	git clone https://github.com/ispras/michman.git
	cd michman

Then, fill Michmans configuration file *config.yaml* with your settings.

.. note:: You can read more about Michman configuration in next section.

Following instructions are given for Michman over Ubuntu18.04 image. If you want to use Centos8 images, replace **Ubuntu18.04** folder in path with the **Centos8** folder.

**Docker Instruction**

	1. Build and run michman-rest service.

	.. parsed-literal::
		docker build -t michman-rest -f ./rest/Ubuntu18.04/Dockerfile .
		docker run --env CONFIG=/config.yaml --env LAUNCHER=localhost:5000 --env PORT=8081 -p 8081:8081 -v <path_to_config>/config.yaml:/config.yaml michman-rest

	2. Build and run michman-launcher service.

	.. parsed-literal::
		docker build -t michman-launcher -f ./launcher/Ubuntu18.04/Dockerfile .
		docker run -p 5000:5000 --env CONFIG=/config.yaml --env PORT=5000 -v <path_to_config>/config.yaml:/config.yaml michman-launcher

**Podman Instruction**

	
	1. Build and run michman-rest service.

	.. parsed-literal::
		podman build -t michman-rest -f ./rest/Ubuntu18.04/Dockerfile .
		podman run --env CONFIG=/config.yaml --env LAUNCHER=localhost:5000 --env PORT=8081 -p 8081:8081 -v <path_to_config>/config.yaml:/config.yaml:z michman-rest

	2. Build and run michman-launcher service.

	.. parsed-literal::
		podman build -t michman-launcher -f ./launcher/Ubuntu18.04/Dockerfile .
		podman run -p 5000:5000 --env CONFIG=/config.yaml --env PORT=5000 -v <path_to_config>/config.yaml:/config.yaml:z michman-launcher



