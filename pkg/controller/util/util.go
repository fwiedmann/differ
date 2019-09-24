package util

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
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

func AddToSortedImages(scrapedImage string, list map[string][]string) {

	registryURL, image := extractScrapedImage(scrapedImage)
	if _, found := list[registryURL]; !found {
		list[registryURL] = make([]string, 0)
	}
	list[registryURL] = append(list[registryURL], image)
}

func extractScrapedImage(image string) (string, string) {
	var host string
	split := strings.Split(image, "/")
	if !strings.Contains(split[0], ".") {
		return "", image
	}

	host = split[0]

	return host, strings.TrimPrefix(image, host+"/")
}
