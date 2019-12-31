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
	log "github.com/sirupsen/logrus"

	"github.com/fwiedmann/differ/pkg/observer/util"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/watch"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fwiedmann/differ/pkg/types"
)

const apiVersion = "appV1"
const deploymentResourceType = "Deployment"

// Deployment types struct
type Deployment struct {
	apiWatcher           watch.Interface
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
	apiClient, namespace := observeConfig.KubernetesAPI.GetAPIClientAndNameSpace()

	if d.apiWatcher, err = apiClient.AppsV1().Deployments(namespace).Watch(metaV1.ListOptions{}); err != nil {
		return err
	}
	d.observeConfig = observeConfig

	return nil
}

func (d *Deployment) UpdateConfig(observeConfig *types.KubernetesObserverConfig) error {
	d.stopObservingChannel <- struct{}{}
	d.apiWatcher.Stop()

	if err := d.InitObserverWithConfig(observeConfig); err != nil {
		return err
	}
	return nil
}

// SendObservedKubernetesAPIResource
func (d *Deployment) SendObservedKubernetesAPIResource() {
	apiObserveChannel := d.apiWatcher.ResultChan()
Loop:
	for {
		select {
		case apiObjectEvent := <-apiObserveChannel:
			parsedDeployment, ok := apiObjectEvent.Object.(*v1.Deployment)
			if !ok {
				log.Info("could not convert event")
				continue
			}

			extractedImagesWithAssociatedPullSecrets, err := util.GetImagesWithImagePullSecrets(parsedDeployment.Spec.Template, d.observeConfig)
			if err != nil {
				log.Info("%s", err.Error())
			}

			kubernetesResourceMetaInfo := types.KubernetesAPIObjectMetaInformation{
				APIVersion:   apiVersion,
				ResourceType: deploymentResourceType,
				Namespace:    parsedDeployment.Namespace,
				WorkloadName: parsedDeployment.Name,
			}

			eventsToSend := util.GenerateEventForEachExtractedImageWithKubernetesMetaDataInfo(apiObjectEvent.Type, extractedImagesWithAssociatedPullSecrets, kubernetesResourceMetaInfo)

			for _, event := range eventsToSend {
				d.observeConfig.ObserverChannel <- event
			}

		case <-d.stopObservingChannel:
			break Loop
		}
	}
}
