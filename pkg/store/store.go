package store

import "strings"

const dockerHubURL string = "https://index.docker.io"

type (
	// Cache stores all scraped images with ResourceMetaInfo in the way of:
	// ["registry"]["image"][]ResourceMetaInfo{}
	Cache map[string]map[string][]ResourceMetaInfo

	// ResourceMetaInfo contains unique meta information from scraped resource types
	ResourceMetaInfo struct {
		APIVersion   string
		ResourceType string
		Namespace    string
		Name         string
		ImageTag     string
	}
)

// AddResource add new resource information to store
func (store Cache) AddResource(scrapedImage, apiVersion, resourceType, namespace, name string) {
	registry, image, tag := getResourceStoreKeys(scrapedImage)

	if _, found := store[registry]; !found {
		store[registry] = make(map[string][]ResourceMetaInfo)
	}

	if _, found := store[registry][image]; !found {
		store[registry][image] = make([]ResourceMetaInfo, 0)
	}
	store[registry][image] = append(store[registry][image], ResourceMetaInfo{
		APIVersion:   apiVersion,
		ResourceType: resourceType,
		Namespace:    namespace,
		Name:         name,
		ImageTag:     tag,
	})
}

// getResourceStoreKeys extract registryURL and image name from scraped image
// If image belongs to docker hub URL will be set to dockerHubURL const
func getResourceStoreKeys(scrapedImage string) (registryURL, image, tag string) {
	split := strings.Split(scrapedImage, "/")
	if !strings.Contains(split[0], ".") {
		image, tag := splitImage(scrapedImage)
		return dockerHubURL, image, tag
	}

	registryURL = split[0]
	trimedImage := strings.TrimPrefix(scrapedImage, registryURL+"/")
	image, tag = splitImage(trimedImage)

	return registryURL, image, tag
}

func splitImage(fullImage string) (image, tag string) {
	split := strings.Split(fullImage, ":")
	if len(split) == 2 {
		return split[0], split[1]
	}
	return fullImage, "latest"
}
