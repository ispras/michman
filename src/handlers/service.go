package handlers

import (
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"log"
)

var servicesTypes = map[string]bool {
	utils.ServiceTypeSpark: true,
	utils.ServiceTypeIgnite: true,
	utils.ServiceTypeJupyter: true,
	utils.ServiceTypeCassandra: true,
	utils.ServiceTypeElastic: true,
	utils.ServiceTypeJupyterhub: true,
}

func ValidateService(service *protobuf.Service) bool {
	if _, ok := servicesTypes[service.Type]; !ok {
		log.Print("ERROR: service type ", service.Type, " is not supported.")
		return false
	}

	//validate config params?

	return true
}

