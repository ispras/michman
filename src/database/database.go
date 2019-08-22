package database

import (
	proto "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
)

type Database interface {
	WriteCluster (cluster *proto.Cluster) error
	ReadCluster (name string) (*proto.Cluster, error)
	ListClusters () []*proto.Cluster
	DeleteCluster (name string) error
}