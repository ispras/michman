package check

import (
	"github.com/ispras/michman/internal/utils"
	"regexp"
)

// ProjectValidName checks the correctness of the project name style for its use
func ProjectValidName(name string) error {
	validName := regexp.MustCompile(utils.ProjectNamePattern).MatchString
	if !validName(name) {
		return ErrClusterBadName
	}
	return nil
}
