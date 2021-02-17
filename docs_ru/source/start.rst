.. _michman_start_section:

Запуск Michman
=================

В этом руководстве описывается, как запустить Michman. Этот раздел будет полезен, если Michman используется не в контейнерах. Иначе Michman запускается автоматически.

Не забудьте заполнить файл *config.yaml*. 

Быстрый старт
--------------

Для быстрого старта рекомендуется использовать скрипт *build.sh*:

.. parsed-literal::
	./build.sh start

Старт в ручном режиме
----------------------

Также вы можете запустить Michman вручную. Сначала сгенерируйте proto-файл: 

.. parsed-literal::
	./build.sh proto

Запустите сервис ansible_runner:

.. parsed-literal::
	go run ./launcher/ansible_launcher.go ./launcher/main.go

Или запустите сервис ansible_runner, указав путь к файлу конфигурации и порт, по умолчанию - устанавливаются путь к файлу конфигурации в корне Michman и 5000 как используемый порт: 

.. parsed-literal::
	go run ./launcher/ansible_launcher.go ./launcher/main.go --config /path/to/config.yaml --port PORT


Запустите http_server:

.. parsed-literal::
	go run ./rest/main.go


Или запустите сервис http_server, указав путь к файлу конфигурации, порт и адрес сервиса *launcher*, по умолчанию - устанавливаются путь к файлу конфигурации в корне Michman, 8081 как используемый порт и localhost:5000 для адреса сервиса *launcher*:

.. parsed-literal::
	go run ./rest/main.go --config /path/to/config.yaml --port PORT --launcher launcher_host:launcher_port

Компиляция при помощи скрипта *build.sh*
-----------------------------------------

Кроме того, вы можете скомпилировать сервисы Michman с помощью скрипта *build.sh*: 

.. parsed-literal::
	./build.sh compile

Далее запустите бинарные файлы. Http-server:

.. parsed-literal::
	./http --config /path/to/config.yaml --port PORT --launcher launcher_host:launcher_port

Launcher:

.. parsed-literal::
	./launch --config /path/to/config.yaml --port PORT
