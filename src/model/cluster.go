package model

type Cluster struct {
	Name   string `json:"name"`
	Slaves int32 `json:"slaves"`
	Status string `json:"status"`
	//	Service []service
}

// func (c *Cluster) getName(){
// 	return c.name
// }

// func (c *Cluster) getStatus(){
// 	return c.status
// }
