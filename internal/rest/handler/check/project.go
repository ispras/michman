package check

import (
	"net/http"
	"regexp"
)

// ValidName checks the correctness of the project name style for its use
func ValidName(name string, pattern string, errorType error) (error, int) {
	validName := regexp.MustCompile(pattern).MatchString
	if !validName(name) {
		return errorType, http.StatusBadRequest
	}
	return nil, 0
}
