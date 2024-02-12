package checker

import (
	"fmt"
	"kube-recall/internal/config"
	"kube-recall/internal/kubernetes"
	"kube-recall/internal/model"
	"kube-recall/internal/provider/datadog"
	"kube-recall/internal/util"
	"sync"
	"time"
)

var logger = util.GetLogger()
var cfg = config.Get()

func CheckResources() {
	metrics, err := datadog.GetMetrics(cfg.Filters)
	if err != nil {
		panic(err.Error())
	}

	var wg sync.WaitGroup

	for namespace, deployments := range calculateNewResources(metrics) {
		for deploy, resources := range deployments {
			wg.Add(1)
			go func(ns string, dp string, res model.Resources) {
				defer wg.Done()
				kubernetes.UpdateDeployment(ns, dp, res)
			}(namespace, deploy, resources)
		}
	}
	wg.Wait()
}

func calculateNewResources(metrics map[string]model.Metrics) map[string]map[string]model.Resources {
	newResources := make(map[string]map[string]model.Resources)
	for _, metric := range metrics {
		avg, maxCpuUsage, maxMemoryUsage := ponderedAverage(metric.TimeSeries)

		newCpuRequest := util.RoundUpToBase10(avg.CPU.Usage * cfg.Scales.Cpu.Request)
		newMemoryRequest := util.RoundUpToBase50(avg.Memory.Usage * cfg.Scales.Memory.Request)
		newCpuLimit := util.RoundUpToBase50(newCpuRequest * cfg.Scales.Cpu.Limit)
		newMemoryLimit := util.RoundUpToBase100(newMemoryRequest * cfg.Scales.Memory.Limit)

		if newCpuLimit < maxCpuUsage {
			newCpuLimit = util.RoundUpToBase50(maxCpuUsage)
		}

		if newMemoryLimit < maxMemoryUsage {
			newMemoryLimit = util.RoundUpToBase100(maxMemoryUsage)
		}
		new := model.Resources{
			Replicas: avg.Replicas,
			CPU: model.MetricFiels{
				Request: newCpuRequest,
				Limit:   newCpuLimit,
			},
			Memory: model.MetricFiels{
				Request: newMemoryRequest,
				Limit:   newMemoryLimit,
			},
		}
		if _, ok := newResources[metric.Namespace]; !ok {
			newResources[metric.Namespace] = make(map[string]model.Resources)
		}
		newResources[metric.Namespace][metric.Deployment] = new

		logger.Debug(fmt.Sprintf("%s => Avg CPU/Memory %f/%f => Max CPU/Memory: %f/%f", metric.Deployment, avg.CPU.Usage, avg.Memory.Usage, maxCpuUsage, maxMemoryUsage))
		logger.Debug(fmt.Sprintf("%s => New values for: CPU: %f/%f, Memory: %f/%f", metric.Deployment, new.CPU.Request, new.CPU.Limit, new.Memory.Request, new.Memory.Limit))
	}

	return newResources
}

func ponderedAverage(resources map[time.Time]model.Resources) (model.Resources, float64, float64) {
	var totalWeight float64
	var cpuUsageSum, cpuRequestSum, cpuLimitSum, memoryUsageSum, memoryRequestSum, memoryLimitSum float64
	var maxCpuUsage, maxMemoryUsage float64 = 0, 0

	for _, resource := range resources {
		totalWeight += resource.Replicas
		cpuUsageSum += resource.CPU.Usage * resource.Replicas
		cpuRequestSum += resource.CPU.Request * resource.Replicas
		cpuLimitSum += resource.CPU.Limit * resource.Replicas
		memoryUsageSum += resource.Memory.Usage * resource.Replicas
		memoryRequestSum += resource.Memory.Request * resource.Replicas
		memoryLimitSum += resource.Memory.Limit * resource.Replicas

		if resource.CPU.Usage > maxCpuUsage {
			maxCpuUsage = resource.CPU.Usage
		}
		if resource.Memory.Usage > maxMemoryUsage {
			maxMemoryUsage = resource.Memory.Usage
		}
	}
	avgCPUUsage := cpuUsageSum / totalWeight
	avgCPURequest := cpuRequestSum / totalWeight
	avgCPULimit := cpuLimitSum / totalWeight
	avgMemoryUsage := memoryUsageSum / totalWeight
	avgMemoryRequest := memoryRequestSum / totalWeight
	avgMemoryLimit := memoryLimitSum / totalWeight

	newResources := model.Resources{
		Replicas: totalWeight,
		CPU: model.MetricFiels{
			Request: avgCPURequest,
			Limit:   avgCPULimit,
			Usage:   avgCPUUsage,
		},
		Memory: model.MetricFiels{
			Request: avgMemoryRequest,
			Limit:   avgMemoryLimit,
			Usage:   avgMemoryUsage,
		},
	}

	return newResources, maxCpuUsage, maxMemoryUsage
}
