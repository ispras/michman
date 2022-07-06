package ansible

import (
	"errors"
	"fmt"
)

const (
	errMarshal          = "error occurred while encoding json"
	errUnMarshal        = "error occurred while parsing JSON"
	errDownload         = "error occurred while downloading file"
	errAbs              = "error occurred while converting path to an absolute representation"
	errConvertParam     = "error occurred while converting params"
	errSetEnv           = "error occurred while setting value of the environment variable named by the key"
	errCmdStart         = "error occurred while starting the specified command"
	errCmdWait          = "error occurred while waiting the command to exit"
	errCouchSecretsRead = "error occurred while reading couchbase secrets"
	errCreate           = "error occurred while creating file"
	errChmod            = "error occurred while changing the mode of the file"
	errWrite            = "error occurred while writing to the file"
	errClose            = "error occurred while closing file"
)

var (
	ErrMarshal          = errors.New(errMarshal)
	ErrDownload         = errors.New(errDownload)
	ErrAbs              = errors.New(errAbs)
	ErrUnMarshal        = errors.New(errUnMarshal)
	ErrConvertParam     = errors.New(errConvertParam)
	ErrSetEnv           = errors.New(errSetEnv)
	ErrCmdStart         = errors.New(errCmdStart)
	ErrCmdWait          = errors.New(errCmdWait)
	ErrCouchSecretsRead = errors.New(errCouchSecretsRead)
	ErrCreate           = errors.New(errCreate)
	ErrChmod            = errors.New(errChmod)
	ErrWrite            = errors.New(errWrite)
	ErrClose            = errors.New(errClose)
)

func ErrParseValue(param string) error {
	ErrParamType := fmt.Errorf("error occurred while parsing %s value", param)
	return ErrParamType
}
