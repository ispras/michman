package logger

import (
	"bytes"
	"encoding/json"
	"github.com/ispras/michman/utils"
	"io"
	"io/ioutil"
	"net/http"
)

type clusterLog struct {
	ClusterName string `json:"Cluster_name"`
	Data        string `json:"Data"`
}

const (
	sqlQueryPath = "/_sql?format=txt"
)

type queryLog struct {
	Query string `json:"query"`
}

type LogstashLogger struct {
	logstashAddr string
	elasticAddr  string
	clusterID    string
	action       string
	logsBuffer   *bytes.Buffer
}

func NewLogstashLogger(clusterID string, action string) (Logger, error) {
	ll := new(LogstashLogger)
	config := utils.Config{}
	config.MakeCfg()
	ll.logstashAddr = config.LogstashAddr
	ll.elasticAddr = config.ElasticAddr
	ll.clusterID = clusterID
	ll.action = action
	ll.logsBuffer = new(bytes.Buffer)
	return ll, nil
}

func makeClusterName(clusterID string, action string) string {
	return action + "_" + clusterID
}

func (ll LogstashLogger) PrepClusterLogsWriter() (io.Writer, error) {
	return ll.logsBuffer, nil
}

func (ll LogstashLogger) FinClusterLogsWriter() error {
	cLog := clusterLog{ClusterName: makeClusterName(ll.clusterID, ll.action), Data: ll.logsBuffer.String()}
	client := http.Client{}

	jsonLogs, err := json.Marshal(cLog)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, ll.logstashAddr, bytes.NewBuffer(jsonLogs))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	client.Do(req)
	return nil
}

func (ll LogstashLogger) ReadClusterLogs() (string, error) {
	getLogQuery := "SELECT * FROM \"" + makeClusterName(ll.clusterID, ll.action) + "\""
	logs := queryLog{Query: getLogQuery}

	client := http.Client{}

	jsonLog, err := json.Marshal(logs)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, ll.elasticAddr+sqlQueryPath, bytes.NewBuffer(jsonLog))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	return bodyString, nil
}
