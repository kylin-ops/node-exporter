package node_exporter

import (
	"encoding/json"
	"fmt"
	"github.com/kylin-ops/node_exporter/collector"
	"github.com/kylin-ops/node_exporter/prometheus/client_golang/prometheus"
	"github.com/kylin-ops/node_exporter/prometheus/client_golang/prometheus/promhttp"
	"github.com/kylin-ops/node_exporter/prometheus/common/log"
	"github.com/kylin-ops/node_exporter/prometheus/common/version"
	"io/ioutil"
	"net/http"
	"sort"
)

var msg = `
	{
		"a": "aa",
		"b": "bb"
	}
`

type Handler struct {
	unfilteredHandler http.Handler
	// exporterMetricsRegistry is a separate registry for the metrics about
	// the exporter itself.
	exporterMetricsRegistry *prometheus.Registry
	includeExporterMetrics  bool
	maxRequests             int
}

func (h *Handler) innerHandler(filters ...string) (http.Handler, error) {
	nc, err := collector.NewNodeCollector(filters...)
	if err != nil {
		return nil, fmt.Errorf("couldn't create collector: %s", err)
	}

	// Only log the creation of an unfiltered handler, which should happen
	// only once upon startup.
	if len(filters) == 0 {
		log.Infof("Enabled collectors:")
		collectors := []string{}
		for n := range nc.Collectors {
			collectors = append(collectors, n)
		}
		sort.Strings(collectors)
		for _, n := range collectors {
			log.Infof(" - %s", n)
		}
	}

	r := prometheus.NewRegistry()
	r.MustRegister(version.NewCollector("node_exporter"))
	if err := r.Register(nc); err != nil {
		return nil, fmt.Errorf("couldn't register node collector: %s", err)
	}
	handler := promhttp.HandlerFor(
		prometheus.Gatherers{h.exporterMetricsRegistry, r},
		promhttp.HandlerOpts{
			ErrorLog:            log.NewErrorLogger(),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: h.maxRequests,
		},
	)
	if h.includeExporterMetrics {
		// Note that we have to use h.exporterMetricsRegistry here to
		// use the same promhttp metrics for all expositions.
		handler = promhttp.InstrumentMetricHandler(
			h.exporterMetricsRegistry, handler,
		)
	}
	return handler, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters := r.URL.Query()["collect[]"]
	log.Debugln("collect query:", filters)

	if len(filters) == 0 {
		// No filters, use the prepared unfiltered handler.
		h.unfilteredHandler.ServeHTTP(w, r)
		return
	}
	// To serve filtered metrics, we create a filtering handler on the fly.
	//filteredHandler, err := h.innerHandler(filters...)
	//if err != nil {
	//	log.Warnln("Couldn't create filtered metrics handler:", err)
	//	w.WriteHeader(http.StatusBadRequest)
	//	w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err)))
	//	return
	//}
	//filteredHandler.ServeHTTP(w, r)
}

func newHandler(includeExporterMetrics bool, maxRequests int) *Handler {
	h := &Handler{
		exporterMetricsRegistry: prometheus.NewRegistry(),
		includeExporterMetrics:  includeExporterMetrics,
		maxRequests:             maxRequests,
	}
	if h.includeExporterMetrics {
		h.exporterMetricsRegistry.MustRegister(
			prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
			prometheus.NewGoCollector(),
		)
	}
	if innerHandler, err := h.innerHandler(); err != nil {
		log.Fatalf("Couldn't create metrics handler: %s", err)
	} else {
		h.unfilteredHandler = innerHandler
	}
	return h
}

func readLabel(labelPath string) error {
	var data = map[string]interface{}{}
	f, err := ioutil.ReadFile(labelPath)
	if err != nil {
		return err
	}
	json.Unmarshal(f, &data)
	service, ok := data["service"]
	if ok {
		if svc, ok := service.([]interface{}); ok {
			for _, svc := range svc {
				if s, ok := svc.(string); ok {
					prometheus.CustomLabelService = append(prometheus.CustomLabelService, s)
				}
			}
		}
	}
	for k, v := range data {
		if k != "service" {
			if val, ok := v.(string); ok {
				prometheus.CustomLabelValue[k] = val
			}
		}
	}
	return nil
}

// 从文件获取label数据
func NewNodeExportHandler(labelPath, scriptPath string) *Handler {
	_ = readLabel(labelPath)
	collector.ScriptPath = scriptPath
	return newHandler(true, 40)
}

// 从函数传入labels和service值
func NewNodeExportHandlerFromValue(labels map[string]string, services []string, scriptPath string) *Handler {
	prometheus.CustomLabelValue = labels
	prometheus.CustomLabelService = services
	collector.ScriptPath = scriptPath
	return newHandler(true, 40)
}
