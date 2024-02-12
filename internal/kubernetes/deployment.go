package kubernetes

import (
	"context"
	"fmt"
	"kube-recall/internal/config"
	"kube-recall/internal/model"
	"strings"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

	updateResources(cliSet, namespace, deployment,
		deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().MilliValue(),
		deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().MilliValue(),
		int64(resources.CPU.Request),
		int64(resources.CPU.Limit),

		deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().MilliValue()/1000/1024/1024,
		deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().MilliValue()/1000/1024/1024,
		int64(resources.Memory.Request),
		int64(resources.Memory.Limit),
	)
	return nil
}

func updateResources(cliSet *kubernetes.Clientset, namespace string, deployment *appv1.Deployment, actualCpuRequest, actualCpuLimit, newCpuRequest, newCpuLimit, actualMemoryRequest, actualMemoryLimit, newMemoryRequest, newMemoryLimit int64) {
	if actualCpuRequest != newCpuRequest || actualCpuLimit != newCpuLimit || actualMemoryRequest != newMemoryRequest || actualMemoryLimit != newMemoryLimit {
		if ok := updateResourceValues(deployment, newCpuRequest, newCpuLimit, newMemoryRequest, newMemoryLimit); ok {
			deploymentUpdated, err := cliSet.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
			if err != nil {
				logger.Error(fmt.Sprintf("Error updating deployment %s in namespace %s: %s", deployment.Name, namespace, err.Error()))
				return
			}

			logger.Info(fmt.Sprintf("Deployment %s in namespace %s updated with new resources: cpu: %vm/%vm => %vm/%vm; memory: %vMi/%vMi => %vMi/%vMi",
				deploymentUpdated.Name, namespace,
				actualCpuRequest, actualCpuLimit, newCpuRequest, newCpuLimit,
				actualMemoryRequest, actualMemoryLimit, newMemoryRequest, newMemoryLimit,
			))
		}
	}
}

func updateResourceValues(deployment *appv1.Deployment, newCpuRequest, newCpuLimit, newMemoryRequest, newMemoryLimit int64) bool {
	needUpdate := false

	if ok := setResourceValue(&deployment.Spec.Template.Spec.Containers[0].Resources.Requests, "cpu", newCpuRequest, resource.NewMilliQuantity); ok && !needUpdate {
		needUpdate = true
	}
	if ok := setResourceValue(&deployment.Spec.Template.Spec.Containers[0].Resources.Limits, "cpu", newCpuLimit, resource.NewMilliQuantity); ok && !needUpdate {
		needUpdate = true
	}
	if ok := setResourceValue(&deployment.Spec.Template.Spec.Containers[0].Resources.Requests, "memory", newMemoryRequest*1024*1024, resource.NewQuantity); ok && !needUpdate {
		needUpdate = true
	}
	if ok := setResourceValue(&deployment.Spec.Template.Spec.Containers[0].Resources.Limits, "memory", newMemoryLimit*1024*1024, resource.NewQuantity); ok && !needUpdate {
		needUpdate = true
	}

	return needUpdate
}

func setResourceValue(resourceList *v1.ResourceList, resourceName string, value int64, resourceFunc func(int64, resource.Format) *resource.Quantity) bool {
	if value > 0 {
		if *resourceList == nil {
			*resourceList = make(v1.ResourceList)
		}
		quantity := resourceFunc(value, resource.DecimalSI)
		(*resourceList)[v1.ResourceName(resourceName)] = *quantity
		return true
	}
	return false
}
