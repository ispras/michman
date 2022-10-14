package utils

import "errors"

// error classes
const (
	UnexpectedError    = 0
	DatabaseError      = 100
	JsonError          = 200
	ValidationError    = 300
	LibError           = 400
	ObjectUsed         = 500
	ObjectExists       = 600
	ObjectUnmodified   = 700
	LogsError          = 800
	AuthorizationError = 900
	ObjectNotFound     = 1000
	InputIncorrect     = 1100
)

const (
	errUnmarshal = "unmarshal error"
	errGetwd     = "error in finding path name for the current directory"
	errMkdir     = "error in creation a new directory"

	errVaultNewClient = "can't create new vault client"
	errVaultReadFile  = "error occurred while reading vault config file"

	errAuthorizationModel      = "for config parameter 'authorization_model' are supported only 'none', 'oauth2' or 'keystone' values"
	errOAuth2ModeAuthorization = "for oauth2 authorization mode config parameters 'hydra_admin' and 'hydra_client' couldn't be empty"
	errKeystoneAuthorization   = "for keystone authorization mode config parameters 'keystone_addr' couldn't be empty"
	errLogsOutputParams        = "for config parameter 'logs_output` are supported only 'file' or 'logstash' values"
	errLogsFilePathEmpty       = "'logs_file_path' couldn't be empty"
	errLogstashOutputParams    = "for logstash logs output config parameters 'logstash_addr' and 'elastic_addr' couldn't be empty"
	errStorage                 = "for storage config parameter are supported only 'couchbase' or 'mysql' values"
)

var (
	ErrVaultNewClient          = errors.New(errVaultNewClient)
	ErrVaultReadFile           = errors.New(errVaultReadFile)
	ErrUnmarshal               = errors.New(errUnmarshal)
	ErrGetwd                   = errors.New(errGetwd)
	ErrAuthorizationModel      = errors.New(errAuthorizationModel)
	ErrOAuth2ModeAuthorization = errors.New(errOAuth2ModeAuthorization)
	ErrKeystoneAuthorization   = errors.New(errKeystoneAuthorization)
	ErrLogsOutputParams        = errors.New(errLogsOutputParams)
	ErrLogsFilePathEmpty       = errors.New(errLogsFilePathEmpty)
	ErrMkdir                   = errors.New(errMkdir)
	ErrLogstashOutputParams    = errors.New(errLogstashOutputParams)
	ErrStorage                 = errors.New(errStorage)
)
