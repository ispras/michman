package logger

import (
	"github.com/ispras/michman/utils"
	"io"
)

type Logger interface {
	PrepClusterLogsWriter() (io.Writer, error)
	FinClusterLogsWriter() error
	ReadClusterLogs() (string, error)
}

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


