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

package controller

import (
	"sync"

	"github.com/fwiedmann/differ/pkg/controller/util"
	"github.com/fwiedmann/differ/pkg/metrics"
	"github.com/fwiedmann/differ/pkg/opts"
	"github.com/fwiedmann/differ/pkg/registry"
	"github.com/fwiedmann/differ/pkg/store"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

//ResourceScraper save scraped data in store
type ResourceScraper interface {
	GetWorkloadResources(c *kubernetes.Clientset, namespace string, scrapedResources *store.Instance) error
}

// Controller type struct
type Controller struct {
	config *opts.ControllerConfig
}

// New initialize the differ controller
func New(c *opts.ControllerConfig) *Controller {
	return &Controller{
		config: c,
	}
}

// Run starts differ controller loop
func (controller *Controller) Run(resourceScrapers []ResourceScraper) error {
	remotes := registry.NewRemoteStore()

	for {
		resourceStore := store.NewInstance()

		kubernetesClient, err := util.InitKubernetesClient()
		if err != nil {
			return err
		}

		for _, s := range resourceScrapers {
			if err := s.GetWorkloadResources(kubernetesClient, controller.config.Namespace, resourceStore); err != nil {
				return err
			}
		}
		metrics.DeleteNotScrapedResources(resourceStore)
		log.Tracef("Scraped resources: %v", resourceStore)

		// limit concurrent execution to 300 with tokens
		workerTokens := make(chan struct{}, 300)
		workerErrors := make(chan error, resourceStore.Size())
		var wg sync.WaitGroup

		// start worker for each image
		for image, imageInfos := range resourceStore.GetDeepCopy() {
			wg.Add(1)
			go func(imageName string, resourceMetaInfos []store.ResourceMetaInfo, errChan chan<- error) {
				workerTokens <- struct{}{}
				defer wg.Done()
				auths := util.GatherAuths(resourceMetaInfos)

				if err := remotes.CreateOrUpdateRemote(imageName, auths); err != nil {
					errChan <- err
				} else {

					remote := remotes.GetRemoteByID(imageName)
					remoteTags, err := remote.GetTags()
					if err != nil {
						errChan <- err
					} else {
						for _, info := range resourceMetaInfos {
							metrics.SetGaugeValueWithID("differ_scraped_image", info.ImageName, info.ImageTag, 1, info.ImageName, info.ImageTag, info.ResourceType, info.WorkloadName, info.APIVersion, info.Namespace)
							valid, pattern := util.IsValidTag(info.ImageTag)
							if !valid {
								continue
							}
							sortedTags := util.SortTagsByPattern(remoteTags, pattern)
							if sortedTags[len(sortedTags)-1] != info.ImageTag {
								metrics.SetGaugeValueWithID("differ_update_image", info.ImageName, info.ImageTag, 1, info.ImageName, info.ImageTag, info.ResourceType, info.WorkloadName, info.APIVersion, info.Namespace, sortedTags[len(sortedTags)-1])
							}
						}
						errChan <- nil
					}
				}
				<-workerTokens
			}(image, imageInfos, workerErrors)
			log.Debugf("%+v", resourceStore)
		}

		// wait for all workers
		go func() {
			wg.Wait()
			close(workerTokens)
			close(workerErrors)
		}()

		for workerError := range workerErrors {
			if err := util.IsRegistryError(workerError); err != nil {
				return err
			}
		}

		controller.config.ControllerSleep()
	}
}
