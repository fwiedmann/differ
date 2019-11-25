package store

import (
	"reflect"
	"strings"
	"sync"

	v1 "k8s.io/api/core/v1"
)

// todo: delete not given resources, to put instance at top level of controller.Run()

type (
	// Cache stores all scraped images with ResourceMetaInfo in the way of:
	// ["image"][]ResourceMetaInfo{}
	Instance struct {
		data map[string][]ResourceMetaInfo
		m    sync.RWMutex
	}

	ImagePullSecret struct {
		Username string
		Password string
	}
	// ResourceMetaInfo contains unique meta information from scraped resource types
	ResourceMetaInfo struct {
		APIVersion   string
		ResourceType string
		Namespace    string
		WorkloadName string
		ImageName    string
		ImageTag     string
		Secrets      []ImagePullSecret
	}
)

func NewInstance() Instance {
	return Instance{
		data: make(map[string][]ResourceMetaInfo, 0),
		m:    sync.RWMutex{},
	}
}

// AddResource add new resource information to store
func (storeInstance Instance) AddResource(apiVersion, kind, namespace, name string, containers []v1.Container, secrets map[string][]ImagePullSecret) {
	storeInstance.m.Lock()
	defer storeInstance.m.Unlock()
	for _, container := range containers {

		image, tag := getResourceStoreKeys(container.Image)

		if _, found := storeInstance.data[image]; !found {
			storeInstance.data[image] = make([]ResourceMetaInfo, 0)
		}

		var matchingSecrets []ImagePullSecret
		for registry, regsistrySecrets := range secrets {
			// consider images from dockerHub, they do not contain any dots
			if strings.Contains(image, registry) || !strings.Contains(image, ".") {
				matchingSecrets = append(matchingSecrets, regsistrySecrets...)
			}
		}
		resourceInfo := ResourceMetaInfo{
			APIVersion:   apiVersion,
			ResourceType: kind,
			Namespace:    namespace,
			WorkloadName: name,
			ImageName:    image,
			ImageTag:     tag,
			Secrets:      matchingSecrets,
		}
		var exists bool
		if len(storeInstance.data[image]) > 0 {
			for _, existingImage := range storeInstance.data[image] {
				if reflect.DeepEqual(existingImage, resourceInfo) {
					exists = true
				}
			}

		}
		if !exists {
			storeInstance.data[image] = append(storeInstance.data[image], resourceInfo)
		}
	}
}

func (storeInstance Instance) GetDeepCopy() map[string][]ResourceMetaInfo {

	storeInstance.m.RLock()
	defer storeInstance.m.RUnlock()
	deepCopy := make(map[string][]ResourceMetaInfo)
	for key, value := range storeInstance.data {
		deepCopy[key] = value
	}
	return deepCopy
}

// Size return current scraped image count
func (storeInstance Instance) Size() int {
	storeInstance.m.RLock()
	defer storeInstance.m.RUnlock()
	return len(storeInstance.data)
}

// getResourceStoreKeys extract image and tag from scraped image
func getResourceStoreKeys(scrapedImage string) (image, tag string) {
	split := strings.Split(scrapedImage, ":")
	if len(split) == 2 {
		return split[0], split[1]
	}
	return scrapedImage, "latest"
}
