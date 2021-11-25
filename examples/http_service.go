package main

import (
	"github.com/kylin-ops/node-exporter"
	"net/http"
)

func main() {
	var (
		listenAddress = ":9100"
		metricsPath   = "/metrics"
	)
	scriptPath := "/tmp/scripts"
	handler := node_exporter.NewNodeExportHandler(scriptPath)
	http.Handle(metricsPath, handler)
	http.ListenAndServe(listenAddress, nil)
}
