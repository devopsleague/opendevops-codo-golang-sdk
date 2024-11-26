package xvm

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ccheers/xpkg/generic/arrayx"
)

// MetricResultStatus 定义查询结果状态类型
type MetricResultStatus string

// ResultType 定义结果类型
type ResultType string

const (
	// MetricResultStatus constants

	MetricResultStatusSuccess MetricResultStatus = "success"
	MetricResultStatusError   MetricResultStatus = "error"

	// ResultType constants

	// ResultTypeMatrix Range Vector (范围向量)
	// 示例查询
	//query := `node_cpu_seconds_total`
	//
	// 返回格式示例
	//{
	//  "resultType": "vector",
	//  "result": [
	//    {
	//      "metric": {"instance": "localhost:9100", "cpu": "0", "mode": "idle"},
	//      "value": [1634567890, "123.45"]
	//    },
	//    {
	//      "metric": {"instance": "localhost:9100", "cpu": "1", "mode": "idle"},
	//      "value": [1634567890, "234.56"]
	//    }
	//  ]
	//}
	ResultTypeMatrix ResultType = "matrix"

	// ResultTypeVector Instant Vector (瞬时向量)
	// 示例查询
	//query := `rate(node_cpu_seconds_total[5m])`
	//
	// 返回格式示例
	//{
	//  "resultType": "matrix",
	//  "result": [
	//    {
	//      "metric": {"instance": "localhost:9100", "cpu": "0"},
	//      "values": [
	//        [1634567890, "123.45"],
	//        [1634567900, "123.46"],
	//        [1634567910, "123.47"]
	//      ]
	//    }
	//  ]
	//}
	ResultTypeVector ResultType = "vector"

	// ResultTypeScalar Scalar (标量)
	// 示例查询
	//query := `count(up)`
	//
	// 返回格式示例
	//{
	//  "resultType": "scalar",
	//  "result": [1634567890, "42"]
	//}
	ResultTypeScalar ResultType = "scalar"

	// ResultTypeString String (字符串)
	// 示例查询
	//query := `prometheus_build_info`
	//
	// 返回格式示例
	//{
	//  "resultType": "string",
	//  "result": [1634567890, "v2.30.0"]
	//}
	ResultTypeString ResultType = "string"
)

// Point 表示时间序列中的一个数据点
type Point [2]interface{}

// Timestamp 获取数据点的时间戳
func (p Point) Timestamp() time.Time {
	// 处理 float64 和 int64 类型的时间戳
	switch ts := p[0].(type) {
	case float64:
		return time.Unix(int64(ts), 0)
	case int64:
		return time.Unix(ts, 0)
	default:
		return time.Time{}
	}
}

// Value 获取数据点的值
func (p Point) Value() float64 {
	// 处理字符串类型的值
	switch v := p[1].(type) {
	case string:
		f, _ := parseFloat(v)
		return f
	case float64:
		return v
	case float32:
		return float64(v)
	default:
		return 0
	}
}

// StringValue 获取数据点的原始字符串值
func (p Point) StringValue() string {
	if str, ok := p[1].(string); ok {
		return str
	}
	return fmt.Sprintf("%v", p[1])
}

// parseFloat 解析字符串为 float64
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// Metric 表示指标数据
type Metric struct {
	// 指标的标签集合
	Labels map[string]string `json:"metric"`
	// matrix 类型的时间序列数据
	Values []Point `json:"values"`
	// vector/scalar 类型的单个值
	Value *Point `json:"value"`
}

// MetricResult 表示查询结果
type MetricResult struct {
	Status    MetricResultStatus `json:"status"`
	IsPartial bool               `json:"isPartial"`
	Data      MetricData         `json:"data"`
	Error     string             `json:"error,omitempty"`
	ErrorType string             `json:"errorType,omitempty"`
}

// MetricData 表示指标数据
type MetricData struct {
	ResultType ResultType `json:"resultType"`
	Result     []Metric   `json:"result"`
}

func (x *MetricData) IsMatrix() bool {
	return x.ResultType == ResultTypeMatrix
}
func (x *MetricData) IsVector() bool {
	return x.ResultType == ResultTypeVector
}
func (x *MetricData) IsScalar() bool {
	return x.ResultType == ResultTypeScalar
}
func (x *MetricData) IsString() bool {
	return x.ResultType == ResultTypeString
}

type MatrixMetric struct {
	// 指标的标签集合
	Labels map[string]string `json:"metric"`
	// matrix 类型的时间序列数据
	Values []Point `json:"values"`
}
type VectorMetric struct {
	// 指标的标签集合
	Labels map[string]string `json:"metric"`
	// vector 类型的时间序列数据
	Value Point `json:"values"`
}
type ScalarMetric struct {
	Value Point `json:"values"`
}
type StringMetric struct {
	Value Point `json:"values"`
}

func (x *MetricData) MatrixValue() []MatrixMetric {
	if !x.IsMatrix() {
		return nil
	}
	return arrayx.Map(x.Result, func(t Metric) MatrixMetric {
		return MatrixMetric{
			Labels: t.Labels,
			Values: t.Values,
		}
	})
}
func (x *MetricData) VectorValue() []VectorMetric {
	if !x.IsVector() {
		return nil
	}
	return arrayx.Map(x.Result, func(t Metric) VectorMetric {
		return VectorMetric{
			Labels: t.Labels,
			Value:  *t.Value,
		}
	})
}
func (x *MetricData) ScalarValue() []ScalarMetric {
	if !x.IsScalar() {
		return nil
	}
	return arrayx.Map(x.Result, func(t Metric) ScalarMetric {
		return ScalarMetric{
			Value: *t.Value,
		}
	})
}
func (x *MetricData) StringValue() []StringMetric {
	if !x.IsString() {
		return nil
	}
	return arrayx.Map(x.Result, func(t Metric) StringMetric {
		return StringMetric{
			Value: *t.Value,
		}
	})
}

// IsSuccess 检查查询是否成功
func (r *MetricResult) IsSuccess() bool {
	return r.Status == MetricResultStatusSuccess
}

// IsError 检查查询是否出错
func (r *MetricResult) IsError() bool {
	return r.Status == MetricResultStatusError
}

// GetValuesForMetric 获取指定指标的所有值
func (r *MetricResult) GetValuesForMetric(metricLabels map[string]string) []Point {
	if !r.IsSuccess() {
		return nil
	}

	for _, metric := range r.Data.Result {
		if matchLabels(metric.Labels, metricLabels) {
			return metric.Values
		}
	}
	return nil
}

// matchLabels 检查标签是否匹配
func matchLabels(actual, expected map[string]string) bool {
	for k, v := range expected {
		if actual[k] != v {
			return false
		}
	}
	return true
}

// exampleUsage Example usage
func exampleUsage() {
	// JSON 示例
	jsonData := `{
        "status": "success",
        "isPartial": false,
        "data": {
            "resultType": "matrix",
            "result": [
                {
                    "metric": {
                        "instance": "10.0.127.118:9100"
                    },
                    "values": [
                        [1732002430, "0.21027777788953017"],
                        [1732002490, "0.21472222223465565"]
                    ]
                }
            ]
        }
    }`

	// 解析 JSON
	var result MetricResult
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		fmt.Printf("解析错误: %v\n", err)
		return
	}

	// 使用解析后的数据
	if result.IsSuccess() {
		for _, metric := range result.Data.Result {
			fmt.Printf("Instance: %s\n", metric.Labels["instance"])
			for _, point := range metric.Values {
				fmt.Printf("Time: %v, Value: %v\n",
					point.Timestamp().Format(time.RFC3339),
					point.Value())
			}
		}
	}
}

// MetricValue 用于格式化输出的辅助结构体
type MetricValue struct {
	Timestamp time.Time
	Value     float64
	RawValue  string
}

// FormatMetricValues 将 Points 转换为更易使用的格式
func FormatMetricValues(points []Point) []MetricValue {
	values := make([]MetricValue, len(points))
	for i, point := range points {
		values[i] = MetricValue{
			Timestamp: point.Timestamp(),
			Value:     point.Value(),
			RawValue:  point.StringValue(),
		}
	}
	return values
}

// GetMetricSeries 获取完整的时间序列数据
func (r *MetricResult) GetMetricSeries() []struct {
	Labels map[string]string
	Values []MetricValue
} {
	if !r.IsSuccess() {
		return nil
	}

	series := make([]struct {
		Labels map[string]string
		Values []MetricValue
	}, len(r.Data.Result))

	for i, metric := range r.Data.Result {
		series[i].Labels = metric.Labels
		series[i].Values = FormatMetricValues(metric.Values)
	}

	return series
}
