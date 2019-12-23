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

package appv1scraper

import (
	"fmt"

	"github.com/fwiedmann/differ/pkg/store"

	"github.com/fwiedmann/differ/pkg/kubernetesscraper/util"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/watch"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fwiedmann/differ/pkg/types"
)

// Deployment types struct
type Deployment struct {
	deploymentAPIWatcher watch.Interface
	stopObservingChannel chan struct{}
	observeConfig        *types.KubernetesObserverConfig
}

func NewDeploymentObserver() *Deployment {
	return &Deployment{
		stopObservingChannel: make(chan struct{}),
	}
}

func (d *Deployment) InitObserverWithConfig(observeConfig *types.KubernetesObserverConfig) error {
	var err error
	if d.deploymentAPIWatcher, err = observeConfig.KubernetesAPIClient.AppsV1().Deployments(observeConfig.NamespaceToScrape).Watch(metaV1.ListOptions{}); err != nil {
		return err
	}
	d.observeConfig = observeConfig

	return nil
}

func (d *Deployment) UpdateConfig(observeConfig *types.KubernetesObserverConfig) error {
	d.stopObservingChannel <- struct{}{}
	d.deploymentAPIWatcher.Stop()

	if err := d.InitObserverWithConfig(observeConfig); err != nil {
		return err
	}
	return nil
}

// SendObservedKubernetesAPIResource
func (d *Deployment) SendObservedKubernetesAPIResource() {
	apiObserveChannel := d.deploymentAPIWatcher.ResultChan()
	for {
		select {
		case apiObjectEvent := <-apiObserveChannel:
			parsedDeployment := apiObjectEvent.Object.(*v1.Deployment)
			extractedImagesWithAssociatedPullSecrets, _ := util.GetImagesAndImagePullSecrets(parsedDeployment.Spec.Template, d.observeConfig)
			eventType := fmt.Sprintf("%s", apiObjectEvent.Type)

			for _, imageWithAssociatedPullSecrets := range extractedImagesWithAssociatedPullSecrets {
				d.observeConfig.ObserverChannel <- types.ObservedImageEvent{
					EventType: eventType,
					ImageWithKubernetesMetadata: store.KubernetesAPIResource{
						APIVersion:   "appV1",
						ResourceType: "Deployment",
						Namespace:    d.observeConfig.NamespaceToScrape,
						WorkloadName: parsedDeployment.Name,
						ImageName:    imageWithAssociatedPullSecrets.ImageName,
						ImageTag:     imageWithAssociatedPullSecrets.ImageTag,
						Secrets:      imageWithAssociatedPullSecrets.ImagePullSecrets,
					},
				}
			}
		case <-d.stopObservingChannel:
			break
		}
	}
}
