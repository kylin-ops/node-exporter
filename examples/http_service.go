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
	// node_exporter.UpdateCustomLabel(map[string]string{"a": "aa"})
	// node_exporter.UpdateServiceLabel([]string{"svc1", "svc2"})
	http.Handle(metricsPath, handler)
	http.ListenAndServe(listenAddress, nil)
}
