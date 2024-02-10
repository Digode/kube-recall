package kubernetes

import (
	"context"
	"fmt"
	"k8s-resources-update/internal/config"
	"k8s-resources-update/internal/model"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func UpdateDeployment(namespace string, deploy string, resources model.Resources) error {
	for _, target := range getTagets() {
		err := updateDeployment(target, namespace, deploy, resources)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateDeployment(target config.Target, namespace string, deploy string, resources model.Resources) error {
	if target.From > "" && target.To > "" {
		deploy = strings.ReplaceAll(deploy, target.From, target.To)
	}
	cliSet, _ := getKubernetesConfig(target)
	deployment, err := cliSet.AppsV1().Deployments(namespace).Get(context.Background(), deploy, metav1.GetOptions{})
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting deployment %s in namespace %s: %s", deploy, namespace, err.Error()))
		return err
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

		cpuRequest := resource.NewMilliQuantity(newCpuRequest, resource.DecimalSI)
		cpuLimit := resource.NewMilliQuantity(newCpuLimit, resource.DecimalSI)
		memoryRequest := resource.NewQuantity(newMemoryRequest*1024*1024, resource.BinarySI)
		memoryLimit := resource.NewQuantity(newMemoryLimit*1024*1024, resource.BinarySI)

		if deployment.Spec.Template.Spec.Containers[0].Resources.Requests == nil {
			deployment.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{}
		}

		if deployment.Spec.Template.Spec.Containers[0].Resources.Requests == nil {
			deployment.Spec.Template.Spec.Containers[0].Resources.Requests = make(v1.ResourceList)
		}

		if deployment.Spec.Template.Spec.Containers[0].Resources.Limits == nil {
			deployment.Spec.Template.Spec.Containers[0].Resources.Limits = make(v1.ResourceList)
		}

		deployment.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"] = *cpuRequest
		deployment.Spec.Template.Spec.Containers[0].Resources.Limits["cpu"] = *cpuLimit
		deployment.Spec.Template.Spec.Containers[0].Resources.Requests["memory"] = *memoryRequest
		deployment.Spec.Template.Spec.Containers[0].Resources.Limits["memory"] = *memoryLimit

		deploymentUpdated, err := cliSet.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
		if err != nil {
			logger.Error(fmt.Sprintf("Error updating deployment %s in namespace %s: %s", deployment.Name, namespace, err.Error()))
		}

		logger.Info(fmt.Sprintf("Deployment %s in namespace %s updated with new resources: cpu: %vm/%vm => %vm/%vm; memory: %vMi/%vMi => %vMi/%vMi",
			deploymentUpdated.Name, namespace,
			actualCpuRequest, actualCpuLimit, newCpuRequest, newCpuLimit,
			actualMemoryRequest, actualMemoryLimit, newMemoryRequest, newMemoryLimit,
		))
	} else {
		logger.Debug(fmt.Sprintf("Deployment %s in namespace %s already has the correct resources", deployment.Name, namespace))
	}
	return nil
}
