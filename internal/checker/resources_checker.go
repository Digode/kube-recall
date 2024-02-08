package checker

import (
	"context"
	"fmt"
	"k8s-resources-update/internal/config"
	"k8s-resources-update/internal/datadog"
	"k8s-resources-update/internal/model"
	"k8s-resources-update/internal/util"
	"strings"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var logger = util.GetLogger()
var cfg = config.Get()

func CheckResources() {
	end := time.Now()
	begin := end.AddDate(0, 0, cfg.DataDog.Times.Begin)
	end = end.AddDate(0, 0, cfg.DataDog.Times.End)
	metrics, err := datadog.GetMetrics(cfg.Filters, begin, end)
	if err != nil {
		panic(err.Error())
	}
	newResources := calculateNewResources(metrics)
	//config, err := rest.InClusterConfig()

	for _, target := range cfg.Kubernetes.Targets {
		config, err := clientcmd.BuildConfigFromFlags("", target.ConfigPath)
		if err != nil {
			panic(err.Error())
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		for namespace, deployments := range newResources {
			for deploy, resources := range deployments {
				if target.From > "" && target.To > "" {
					deploy = strings.ReplaceAll(deploy, target.From, target.To)
				}

				deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.Background(), deploy, v1.GetOptions{})
				if err != nil {
					logger.Error(fmt.Sprintf("Error getting deployment %s in namespace %s: %s", deploy, namespace, err.Error()))
				}

				actualCpuRequest := deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().MilliValue()
				actualCpuLimit := deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()
				actualMemoryRequest := deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().MilliValue() / 1000 / 1024 / 1024
				actualMemoryLimit := deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().MilliValue() / 1000 / 1024 / 1024

				newCpuRequest := int64(resources.CPU.Request)
				newCpuLimit := int64(resources.CPU.Limit)
				newMemoryRequest := int64(resources.Memory.Request)
				newMemoryLimit := int64(resources.Memory.Limit)

				if actualCpuRequest != newCpuRequest ||
					actualCpuLimit != newCpuLimit ||
					actualMemoryRequest != newMemoryRequest ||
					actualMemoryLimit != newMemoryLimit {

					deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().SetMilli(int64(resources.CPU.Request))
					deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().SetMilli(int64(resources.CPU.Limit))
					deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().SetMilli(int64(resources.Memory.Request))
					deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().SetMilli(int64(resources.Memory.Limit))

					_, err = clientset.AppsV1().Deployments(namespace).Update(context.Background(), deployment, v1.UpdateOptions{})
					if err != nil {
						logger.Error(fmt.Sprintf("Error updating deployment %s in namespace %s: %s", deployment.Name, namespace, err.Error()))
					}
				} else {
					logger.Info(fmt.Sprintf("Deployment %s in namespace %s already has the correct resources", deployment.Name, namespace))
				}
			}
		}
	}
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

		logger.Info(fmt.Sprintf("%s => Avg CPU/Memory %f/%f => Max CPU/Memory: %f/%f", metric.Deployment, avg.CPU.Usage, avg.Memory.Usage, maxCpuUsage, maxMemoryUsage))
		logger.Info(fmt.Sprintf("%s => New values for: CPU: %f/%f, Memory: %f/%f", metric.Deployment, new.CPU.Request, new.CPU.Limit, new.Memory.Request, new.Memory.Limit))

		logger.Info("-----------------------------------")
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
