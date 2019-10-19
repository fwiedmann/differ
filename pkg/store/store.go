package store

import "strings"

const dockerHubURL string = "https://index.docker.io/"

type (
	// Cache stores all scraped images with ResourceMetaInfo in the way of:
	// ["image"][]ResourceMetaInfo{}
	Cache map[string][]ResourceMetaInfo

	// ResourceMetaInfo contains unique meta information from scraped resource types
	ResourceMetaInfo struct {
		APIVersion   string
		ResourceType string
		Namespace    string
		WorkloadName string
		ImageName    string
		ImageTag     string
	}
)

// AddResource add new resource information to store
func (store Cache) AddResource(scrapedImage, apiVersion, resourceType, namespace, name string) {
	image, tag := getResourceStoreKeys(scrapedImage)

	if _, found := store[image]; !found {
		store[image] = make([]ResourceMetaInfo, 0)
	}

	store[image] = append(store[image], ResourceMetaInfo{
		APIVersion:   apiVersion,
		ResourceType: resourceType,
		Namespace:    namespace,
		WorkloadName: name,
		ImageName:    image,
		ImageTag:     tag,
	})
}

// getResourceStoreKeys extract image and tag from scraped image
// If image belongs to docker hub URL will be set to dockerHubURL const
func getResourceStoreKeys(scrapedImage string) (image, tag string) {
	split := strings.Split(scrapedImage, "/")
	if !strings.Contains(split[0], ".") {
		image, tag := splitImage(scrapedImage)
		return dockerHubURL + image, tag
	}

	image, tag = splitImage(scrapedImage)

	return image, tag
}

func splitImage(fullImage string) (image, tag string) {
	split := strings.Split(fullImage, ":")
	if len(split) == 2 {
		return split[0], split[1]
	}
	return fullImage, "latest"
}
