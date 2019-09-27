/*
 * MIT License
 *
 * Copyright (c) 2019 Felix Wiedmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

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
