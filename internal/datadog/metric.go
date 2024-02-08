package datadog

import (
	"fmt"
	"k8s-resources-update/internal/model"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

func GetMetrics(filters map[string]string, begin, end time.Time) (map[string]model.Metrics, error) {
	beginUnix := begin.Unix()
	endUnix := end.Unix()
	query := buildCompleteQuery(filters)
	allMetrics, err := getMetrics(beginUnix, endUnix, query)
	if err != nil {
		return nil, err
	}
	return mapMetrics(allMetrics), nil
}

func buildCompleteQuery(filters map[string]string) string {
	filterDeploy := buildFilterQuery(filters)
	metrics := []string{
		cfg.DataDog.Tags.Replicas,
		cfg.DataDog.Tags.Memory.Usage,
		cfg.DataDog.Tags.Memory.Request,
		cfg.DataDog.Tags.Memory.Limit,
		cfg.DataDog.Tags.Cpu.Usage,
		cfg.DataDog.Tags.Cpu.Request,
		cfg.DataDog.Tags.Cpu.Limit,
	}
	metricQuery := "avg:%s{%s} by {%s, %s}"
	queries := make([]string, len(metrics))

	for i, metric := range metrics {
		queries[i] = fmt.Sprintf(metricQuery, metric, filterDeploy, cfg.DataDog.Tags.Namespace, cfg.DataDog.Tags.Deployment)
	}
	return strings.Join(queries, ", ")
}

func buildFilterQuery(filters map[string]string) string {
	parts := make([]string, 0, len(filters))
	for key, val := range filters {
		parts = append(parts, fmt.Sprintf("%s:%s", key, val))
	}
	return strings.Join(parts, ", ")
}

func mapMetrics(allMetrics []model.MetricItem) map[string]model.Metrics {
	result := make(map[string]model.Metrics)
	for _, metric := range allMetrics {
		if _, ok := result[metric.Deployment]; !ok {
			result[metric.Deployment] = model.Metrics{
				Namespace:  metric.Namespace,
				Deployment: metric.Deployment,
				TimeSeries: make(map[time.Time]model.Resources),
			}
		}
		metrics, _ := result[metric.Deployment]
		resource, ok := metrics.TimeSeries[metric.Time]
		if !ok {
			resource = model.Resources{}
		}
		resource = getMetric(metric, resource)
		metrics.TimeSeries[metric.Time] = resource
	}
	return result
}

func getMetric(metric model.MetricItem, resource model.Resources) model.Resources {
	switch metric.Kind {
	case cfg.DataDog.Tags.Replicas:
		resource.Replicas = metric.Value
	case cfg.DataDog.Tags.Memory.Usage:
		resource.Memory.Usage = bytesToMb(metric.Value)
	case cfg.DataDog.Tags.Memory.Request:
		resource.Memory.Request = bytesToMb(metric.Value)
	case cfg.DataDog.Tags.Memory.Limit:
		resource.Memory.Limit = bytesToMb(metric.Value)
	case cfg.DataDog.Tags.Cpu.Usage:
		resource.CPU.Usage = nanoToCore(metric.Value)
	case cfg.DataDog.Tags.Cpu.Request:
		resource.CPU.Request = clockToMilicore(metric.Value)
	case cfg.DataDog.Tags.Cpu.Limit:
		resource.CPU.Limit = clockToMilicore(metric.Value)
	}
	return resource
}

func getMetrics(begin, end int64, query string) ([]model.MetricItem, error) {
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	metricsApi := datadogV1.NewMetricsApi(apiClient)
	resp, _, err := metricsApi.QueryMetrics(getContext(), begin, end, query)
	if err != nil {
		return nil, err
	}

	var metrics []model.MetricItem
	for _, serie := range resp.Series {
		tags := serie.GetTagSet()
		namespace := getTag(tags, cfg.DataDog.Tags.Namespace)
		deployment := getTag(tags, cfg.DataDog.Tags.Deployment)

		if namespace == "" || deployment == "" || deployment == "N/A" {
			continue
		}
		for _, point := range serie.Pointlist {
			if len(point) != 2 || point[1] == nil {
				continue
			}
			stampTime := timeSpampToTime(int64(*point[0]))
			metrics = append(metrics, model.MetricItem{
				Namespace:  namespace,
				Deployment: deployment,
				Kind:       serie.GetMetric(),
				Time:       stampTime,
				Value:      *point[1],
			})
		}
	}
	return metrics, nil
}
