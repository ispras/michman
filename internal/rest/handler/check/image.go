package check

import (
	"github.com/ispras/michman/internal/database"
)

// ImageUsed checks whether the image is used in any of the clusters or projects
func ImageUsed(db database.Database, name string) (bool, error) {
	clusters, err := db.ReadClustersList()
	if err != nil {
		return false, err
	}
	for _, c := range clusters {
		if c.Image == name {
			return true, nil
		}
	}
	projects, err := db.ReadProjectsList()
	if err != nil {
		return false, err
	}
	for _, p := range projects {
		if p.DefaultImage == name {
			return true, nil
		}
	}
	return false, nil
}
