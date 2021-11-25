package collector

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/kylin-ops/node-exporter/prometheus/client_golang/prometheus"
	"github.com/kylin-ops/node-exporter/prometheus/common/log"
	"io/ioutil"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

type customScript struct {
	entries *prometheus.Desc
}

type scriptResult struct {
	Description string            `json:"description"`
	MetricName  string            `json:"metric_name"`
	Labels      map[string]string `json:"labels"`
	ValueType   string            `json:"value_type"`
	Value       float64           `json:"value"`
}

var ScriptPath string

func init() {
	registerCollector("customCollector", defaultEnabled, NewCustomCollector)
}

// NewARPCollector returns a new Collector exposing ARP stats.
func NewCustomCollector() (Collector, error) {
	return &customScript{}, nil
}

func listScript() []string {
	var scripts []string
	fs, _ := ioutil.ReadDir(ScriptPath)
	reg := regexp.MustCompile("^exporter_+.")
	for _, f := range fs {
		if !f.IsDir() {
			if reg.MatchString(f.Name()) {
				scripts = append(scripts, path.Join(ScriptPath, f.Name()))
			}
		}
	}
	return scripts
}

func execScript(script string) (*scriptResult, error) {
	var stdout, stderr bytes.Buffer
	var result = &scriptResult{ValueType: "gauge"}
	cmd := exec.Command("sh", "-c", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	time.AfterFunc(time.Second*2, func() {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	})
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, err
	}
	if !(strings.ToLower(result.ValueType) == "counter" || strings.ToLower(result.ValueType) == "gauge") {
		return nil, errors.New("value_type值只能是counter或gauge")
	}
	if result.MetricName == "" {
		return nil, errors.New("metric_name不能为空")
	}
	return result, nil
}

func (c *customScript) Update(ch chan<- prometheus.Metric) error {
	wg := sync.WaitGroup{}
	for _, script := range listScript() {
		wg.Add(1)
		script := script
		go func() {
			r, e := execScript(script)
			if e != nil {
				log.Error(e)
				wg.Done()
				return
			}
			var labelKey, labelVal []string
			var valueType prometheus.ValueType
			switch strings.ToLower(r.ValueType) {
			case "counter":
				valueType = prometheus.CounterValue
			default:
				valueType = prometheus.GaugeValue
			}
			for k, v := range r.Labels {
				labelKey = append(labelKey, k)
				labelVal = append(labelVal, v)
			}
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(r.MetricName, "customScript", "custom"),
					r.Description,
					labelKey, nil,
				),
				valueType, r.Value, labelVal...,
			)
			wg.Done()
		}()
	}
	wg.Wait()
	return nil
}
