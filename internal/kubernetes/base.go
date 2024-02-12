package kubernetes

import (
	"kube-recall/internal/config"
	"kube-recall/internal/util"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clientset map[string]*kubernetes.Clientset = make(map[string]*kubernetes.Clientset)
	logger                                     = util.GetLogger()
	cfg                                        = config.Get()
)

func getTagets() []config.Target {
	targets := cfg.Kubernetes.Targets
	if len(targets) == 0 {
		targets = append(targets, config.Target{})
	}
	return targets
}

func getKubernetesConfig(target config.Target) (*kubernetes.Clientset, error) {
	if ok := clientset[target.ID()]; ok != nil {
		return clientset[target.ID()], nil
	}
	myClientset := &kubernetes.Clientset{}
	if target.ConfigPath > "" {
		cs, err := buildConfig(target)
		if err != nil {
			return nil, err
		}
		myClientset = cs
	} else {
		cs, err := buildInClusterConfig()
		if err != nil {
			return nil, err
		}
		myClientset = cs
	}
	clientset[target.ID()] = myClientset
	return myClientset, nil
}

func buildInClusterConfig() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func buildConfig(target config.Target) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", target.ConfigPath)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset, nil
}
