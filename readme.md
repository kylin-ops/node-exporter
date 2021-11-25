
# 1 增加功能
## 1.1 读取指定文件，全局给数据打上新的标签
## 1.2 执行自定义脚本，将执行结果生产成新的metric

# 2 使用规则
## 2.1 自定义标记规范
```json
{
   "label1": "value1",
   "label2": "value2",
   "service": ["svc1", "svc2"]
}
```
service 有多少个元素，建返回多位条数据
## 2.2 自定义脚本返回数据规范
```json
{
	"description": "",
	"metric_name": "",
	"labels": {"k1":"v1", "k2":"v2"},
	"value_type": "Counter|Gauge",
	"value": 123456
}
```
## 2.3 自定义脚本命名规范
- 脚本名称必须以"exporter_"开头
- 脚本支持bash和python， 必须按照脚本规范，第一行注释脚本解释器
- metric_name 不能重复

## 2.4 脚本范例
python 脚本
```python
#!/bin/python

print('''{
        "description": "this is test metris",
        "metric_name": "node_this_random_py",
        "value_type": "counter",
        "labels": {"k1":"v1", "k2":"v2"},
        "value": 1.2345
}''')
```
bash 脚本
```shell script
#! /bin/bash
echo '
{
        "description": "this is test metris",
        "metric_name": "node_this_random",
        "value_type": "counter",
        "labels": {"k1":"v1", "k2":"v2"},
        "value": 1.2345
}
'
```

# 3 程序调研启动http服务范例
```go
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
	labelsPath := "/tmp/labels"
	scriptPath := "/tmp/scripts"
	handler := node_exporter.NewNodeExportHandler(labelsPath, scriptPath)
	http.Handle(metricsPath, handler)
	http.ListenAndServe(listenAddress, nil)
}
```