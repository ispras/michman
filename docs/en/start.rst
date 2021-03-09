.. _michman_start_section:

Starting Michman
=================

This guide describes how to start Michman. It is usefull, if you don't use Michman in containers, otherwise Michman starts automatically.

Don't forget to fill the *config.yaml* file.

Quick start
------------

To quick start you may use *build.sh* script:

.. parsed-literal::
	./build.sh start

Start in manual mode
--------------------

Also, you can use mannualy startup. First, build proto-file:

.. parsed-literal::
	./build.sh proto

Manually launch ansible_runner service:

.. parsed-literal::
	go run ./launcher/ansible_launcher.go ./launcher/main.go


Manually launch ansible_runner service specifying config and port, defaults are config path in Michman root and 5000 as used port:

.. parsed-literal::
	go run ./launcher/ansible_launcher.go ./launcher/main.go --config /path/to/config.yaml --port PORT


Manually launch http_server:

.. parsed-literal::
	go run ./rest/main.go


Manually launch http_server specifying config, port and launcher address, defaults are config path in Michman root, 8081 as used port and localhost:5000 for launcher address:

.. parsed-literal::
	go run ./rest/main.go --config /path/to/config.yaml --port PORT --launcher launcher_host:launcher_port

Compilation with *build.sh* script 
--------------------------------------

Furthermore, you can compile michman services with  *build.sh* script:

.. parsed-literal::
	./build.sh compile

Then, launch binary files. Http-server:

.. parsed-literal::
	./http --config /path/to/config.yaml --port PORT --launcher launcher_host:launcher_port

Launcher service:

.. parsed-literal::
	./launch --config /path/to/config.yaml --port PORT
