package logger

import (
	"github.com/ispras/michman/internal/utils"
	"io"
)

type Logger interface {
	PrepClusterLogsWriter() (io.Writer, error)
	FinClusterLogsWriter() error
	ReadClusterLogs() (string, error)
}

// MakeNewClusterLogger create new cluster logger file or logstash logger (depending on 'logs_output' in configuration file)
func MakeNewClusterLogger(cfg utils.Config, clusterID string, action string) (Logger, error) {
	var clusterLogger Logger
	var err error
	if cfg.LogsOutput == utils.LogsFileOutput {
		clusterLogger, err = NewFileLogger(clusterID, action)
		if err != nil {
			return nil, err
		}
	} else if cfg.LogsOutput == utils.LogsLogstashOutput {
		clusterLogger, err = NewLogstashLogger(clusterID, action)
		if err != nil {
			return nil, err
		}
	}
	return clusterLogger, nil
}
