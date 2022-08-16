package cmd

import (
	"errors"
	"fmt"
)

const (
	errOpenFile           = "error occurred while opening file"
	errServe              = "failed to serve"
	errAtoi               = "failed to convert value to int"
	errNewEnforcerSafe    = "error creating new enforcer"
	errAuthModelNotExists = "auth_model.conf does not exist"
	errPolicyNotExists    = "policy.csv does not exist"
)

var (
	ErrOpenFile           = errors.New(errOpenFile)
	ErrServe              = errors.New(errServe)
	ErrAtoi               = errors.New(errAtoi)
	ErrNewEnforcerSafe    = errors.New(errNewEnforcerSafe)
	ErrAuthModelNotExists = errors.New(errAuthModelNotExists)
	ErrPolicyNotExists    = errors.New(errPolicyNotExists)
)

func ErrTcpListen(param string) error {
	ErrParamType := fmt.Errorf("failed to listen: %s", param)
	return ErrParamType
}
