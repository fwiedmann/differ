package scraper

// ScrapedResource type struct
type ScrapedResource struct {
	APIVersion   string
	ResourceType string
	Namespace    string
	Name         string
}

// AddRessource add new resource information to store
func AddRessource(image, apiVersion, resourceType, namespace, name string, store map[string][]ScrapedResource) {
	if _, found := store[image]; !found {
		store[image] = make([]ScrapedResource, 0)
	}
	store[image] = append(store[image], ScrapedResource{
		APIVersion:   apiVersion,
		ResourceType: resourceType,
		Namespace:    namespace,
		Name:         name,
	})
}
