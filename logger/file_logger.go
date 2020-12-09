package logger

import (
	"github.com/ispras/michman/utils"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type FileLogger struct {
	filePath  string
	logFile   *os.File
	clusterID string
	action    string
}

func NewFileLogger(clusterID string, action string) (Logger, error) {
	fl := new(FileLogger)
	config := utils.Config{}
	config.MakeCfg()
	fl.filePath = config.LogsFilePath

	fl.clusterID = clusterID
	fl.action = action

	var err error
	fl.logFile, err = os.OpenFile(fl.makeFileName(clusterID, action), os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return fl, nil
}

func (fl FileLogger) makeFileName(clusterID string, action string) string {
	return fl.filePath + "/" + action + "_" + clusterID + ".log"
}

func (fl FileLogger) PrepClusterLogsWriter() (io.Writer, error) {
	return fl.logFile, nil
}

func (fl FileLogger) FinClusterLogsWriter() error {
	err := fl.logFile.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (fl FileLogger) ReadClusterLogs() (string, error) {
	clusterLogs, err := ioutil.ReadFile(fl.makeFileName(fl.clusterID, fl.action))
	if err != nil {
		return "", err
	}
	err = fl.logFile.Close()
	if err != nil {
		log.Println("Error in closing cluster logs file:")
		log.Println(err)
	}
	return string(clusterLogs), nil
}
