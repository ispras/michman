package main

import (
	"log"

	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
)

type AnsibleLauncher struct{}

func (aL AnsibleLauncher) Run(cluster *protobuf.Cluster) error {
	log.SetPrefix("ANSIBLE_LAUNCHER: ")

	// creating ansible-playbook commands according to cluster object

	log.Print("Launch: OK")
	return nil
}
