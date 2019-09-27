package scraper

import "strings"

type (
	// ResourceStore stores all scraped images with ResourceMetaInfo in the way of:
	// ["registry"]["image"][]ResourceMetaInfo{}
	ResourceStore map[string]map[string][]ResourceMetaInfo

	// ResourceMetaInfo contains unique meta information from scraped resource types
	ResourceMetaInfo struct {
		APIVersion   string
		ResourceType string
		Namespace    string
		Name         string
	}
)

// AddResource add new resource information to store
func AddResource(scrapedImage, apiVersion, resourceType, namespace, name string, store ResourceStore) {
	registry, image := getResourceStoreKeys(scrapedImage)

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
	})
}

// getResourceStoreKeys extract registryURL and image name from scraped image
// If image belongs to hub.docker.io the url is set to empty string
func getResourceStoreKeys(scrapedImage string) (registryURL, image string) {
	split := strings.Split(scrapedImage, "/")
	if !strings.Contains(split[0], ".") {
		return "", scrapedImage
	}

	registryURL = split[0]

	return registryURL, strings.TrimPrefix(scrapedImage, registryURL+"/")
}
