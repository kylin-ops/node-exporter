package main

import (
	"github.com/kylin-ops/node_exporter"
	"net/http"
)

func main() {
	var (
		listenAddress = ":9100"
		metricsPath   = "/metrics"
	)
	labelsPath := "/tmp/labels"
	scriptPath := "/tmp/scripts"
	handler := node_exporter.NewNodeExportHandler(labelsPath, scriptPath)
	http.Handle(metricsPath, handler)
	http.ListenAndServe(listenAddress, nil)
}
