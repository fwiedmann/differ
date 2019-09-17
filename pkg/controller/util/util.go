package util

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func InitKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return &kubernetes.Clientset{}, err
	}

	if clientset, err := kubernetes.NewForConfig(config); err != nil {
		return &kubernetes.Clientset{}, err
	} else {
		return clientset, nil
	}
}
