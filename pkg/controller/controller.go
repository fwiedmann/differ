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
	"github.com/fwiedmann/differ/pkg/observer"
	log "github.com/sirupsen/logrus"

	"github.com/fwiedmann/differ/pkg/event"
)

type Observer interface {
	InitObserverWithKubernetesSharedInformer(observerConfig observer.Config)
	StartObserving()
}

// DifferController types struct
type DifferController struct {
	kubernetesEventChannels event.KubernetesEventCommunicationChannels
	observers               []Observer
}

// NewDifferController initialize the differ controller
func NewDifferController(kubernetesEventChannels event.KubernetesEventCommunicationChannels, observers ...Observer) *DifferController {
	return &DifferController{
		kubernetesEventChannels: kubernetesEventChannels,
		observers:               observers,
	}
}

// StartController starts differ controller loop
func (c *DifferController) StartController() error {
	for _, o := range c.observers {
		go o.StartObserving()
	}
	for {
		select {
		case createEvent := <-c.kubernetesEventChannels.GetADDReceiverEventChanel():
			log.Infof("create event: %+v", createEvent)
		case deleteEvent := <-c.kubernetesEventChannels.GetDELETReceiverEventChanel():
			log.Infof("delete event: %v", deleteEvent)
		case updateEvent := <-c.kubernetesEventChannels.GetUPDATEReceiverEventChanel():
			log.Infof("update event: %v", updateEvent)
		case errorEvent := <-c.kubernetesEventChannels.GetERRORReceiverEventChanel():
			panic(errorEvent)
		}
	}

	/*o, err := observer.InitKubernetesAPIObservers(client, controllerConfig.Namespace)
	if err != nil {
		return err
	}

	o.ObserveAPIs()
	go func() {
		for {
			log.Infof("%+v", <-o.ObserverChannel)
		}
	}()*/
	/*remotes := registry.NewRemoteStore()

		for {
			conf := controller.controllerConfig.GetConfig()
			timeBefore := time.Now()
			resourceStore := store.NewInstance()

			kubernetesClient, err := util.InitKubernetesClient()
			if err != nil {
				return err
			}

			for _, s := range resourceScrapers {
				if err := s.GetWorkloadResources(kubernetesClient, conf.Namespace, resourceStore); err != nil {
					return err
				}
			}
			metrics.DifferConfig.WithLabelValues(conf.Version, conf.Namespace, conf.Sleep, strconv.Itoa(conf.Metrics.Port), conf.Metrics.Path)
			metrics.DifferControllerRuns.Inc()
			metrics.DifferScrapedImages.Set(float64(resourceStore.Size()))
			metrics.DeleteNotScrapedResources(resourceStore)

			log.Tracef("Scraped resources: %v", resourceStore)

			// limit concurrent execution to 300 with tokens
			workerTokens := make(chan struct{}, 300)
			workerErrors := make(chan error, resourceStore.Size())
			var wg sync.WaitGroup

			// start worker for each image
			for image, imageInfos := range resourceStore.GetDeepCopy() {
				wg.Add(1)
				go func(imageName string, resourceMetaInfos []store.KubernetesAPIResource, errChan chan<- error) {
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
								metrics.DynamicMetricSetGaugeValue("differ_scraped_image", 1, info.ImageName, info.ImageTag, info.ResourceType, info.WorkloadName, info.APIVersion, info.Namespace)
								valid, pattern := util.IsValidTag(info.ImageTag)
								if !valid {
									log.Debugf("Tag %s from image %s does not match any valid pattern", info.ImageTag, info.ImageName)
									metrics.DynamicMetricSetGaugeValue("differ_unknown_image_tag", 1, info.ImageName, info.ImageTag, info.ResourceType, info.WorkloadName, info.APIVersion, info.Namespace)
									continue
								}
								sortedTags := util.SortTagsByPattern(remoteTags, pattern)
								if sortedTags[len(sortedTags)-1] != info.ImageTag {
									metrics.DynamicMetricSetGaugeValue("differ_update_image", 1, info.ImageName, info.ImageTag, info.ResourceType, info.WorkloadName, info.APIVersion, info.Namespace, sortedTags[len(sortedTags)-1])
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
	observerConfig
			for workerError := range workerErrors {
				if err := util.IsRegistryError(workerError); err != nil {
					return err
				}
			}

			timeAfter := time.Now()
			metrics.DifferControllerDuration.Set(float64(timeAfter.Sub(timeBefore).Milliseconds()))

			controller.controllerConfig.ControllerSleep()
		}*/

	return nil
}
