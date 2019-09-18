package scraper

func AddNewEntry(containerImage, namespace, resourceType, resourceName string, m map[string]map[string]map[string]string) {
	if _, found := m[containerImage]; !found {
		m[containerImage] = make(map[string]map[string]string)
	}
	if _, found := m[containerImage][namespace]; !found {
		m[containerImage][namespace] = make(map[string]string)
	}
	m[containerImage][namespace][resourceType] = resourceName

}
