package xvm

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestVM(t *testing.T) {

	// 创建带认证的客户端
	client, err := NewMetricsClient(
		os.Getenv("VM_URL"),
		WithClientOptionBasicAuth(os.Getenv("VM_USER"), os.Getenv("VM_PASSWORD")),
	)
	if err != nil {
		panic(err)
	}

	// 设置上下文
	ctx := context.Background()

	// 示例查询
	//
	//// CPU 使用率查询
	//cpuQuery := `100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
	//fmt.Println("查询 CPU 使用率...")
	//result, err := client.QueryRange(ctx, cpuQuery, WithQueryRangeOptionLimit(2))
	//if err != nil {
	//	fmt.Printf("CPU 查询失败: %v\n", err)
	//} else {
	//	PrintMetricResult(result)
	//}
	//
	//// 内存使用率查询
	//memQuery := `100 * (1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)`
	//fmt.Println("\n查询内存使用率...")
	//result, err = client.Query(ctx, memQuery)
	//if err != nil {
	//	fmt.Printf("内存查询失败: %v\n", err)
	//} else {
	//	PrintMetricResult(result)
	//}

	datas := []string{
		//"server_name",
		"entity_count{server_name=\"xxxxx\"}",
		//"online_number",
		//"lock_entity_status",
		//"lock_lb_status",
		//"count(entity_count)",
		//"prometheus_build_info",
	}
	for _, data := range datas {
		result, err := client.Query(ctx, data, WithQueryOptionLimit(1))
		if err != nil {
			fmt.Printf("%s 查询失败: %v\n", data, err)
		} else {
			fmt.Println("\n==============" + data + "==============\n" + "查询结果:\n")
			PrintMetricResult(result)
		}
	}
}

// PrintMetricResult 打印指标查询结果
func PrintMetricResult(result *QueryResult) {
	fmt.Printf("结果类型: %s\n", result.Data.ResultType)
	for _, series := range result.Data.MatrixValue() {
		fmt.Println("\n指标标签:")
		for k, v := range series.Labels {
			fmt.Printf("%s = %s\n", k, v)
		}

		fmt.Println("数据点(Values):")
		for _, point := range series.Values {
			fmt.Printf("%s: %v\n", point.Timestamp().Format(time.RFC3339), point.Value())
		}
	}
	for _, series := range result.Data.VectorValue() {
		fmt.Println("\n指标标签:")
		for k, v := range series.Labels {
			fmt.Printf("%s = %s\n", k, v)
		}

		fmt.Println("数据点(Value):")
		fmt.Printf("%s: %v\n", series.Value.Timestamp().Format(time.RFC3339), series.Value.Value())
	}
}
