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
	"github.com/fwiedmann/differ/pkg/observer/util"
	"github.com/fwiedmann/differ/pkg/types"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const statefulSetResourceType = "StatefulSet"

// StatefulSet types struct
type StatefulSet struct {
	apiWatcher           watch.Interface
	stopObservingChannel chan struct{}
	observeConfig        *types.KubernetesObserverConfig
}

func NewStatefulSetObserver() *StatefulSet {
	return &StatefulSet{
		stopObservingChannel: make(chan struct{}),
	}
}

func (d *StatefulSet) InitObserverWithConfig(observeConfig *types.KubernetesObserverConfig) error {
	var err error
	apiClient, namespace := observeConfig.KubernetesAPI.GetAPIClientAndNameSpace()

	if d.apiWatcher, err = apiClient.AppsV1().StatefulSets(namespace).Watch(metaV1.ListOptions{}); err != nil {
		return err
	}
	d.observeConfig = observeConfig

	return nil
}

func (d *StatefulSet) UpdateConfig(observeConfig *types.KubernetesObserverConfig) error {
	d.stopObservingChannel <- struct{}{}
	d.apiWatcher.Stop()

	if err := d.InitObserverWithConfig(observeConfig); err != nil {
		return err
	}
	return nil
}

// SendObservedKubernetesAPIResource
func (d *StatefulSet) SendObservedKubernetesAPIResource() {
	apiObserveChannel := d.apiWatcher.ResultChan()
Loop:
	for {
		select {
		case apiObjectEvent := <-apiObserveChannel:
			parsedStatefulSet, ok := apiObjectEvent.Object.(*v1.StatefulSet)
			if !ok {
				log.Info("could not convert event")
				continue
			}

			extractedImagesWithAssociatedPullSecrets, err := util.GetImagesWithImagePullSecrets(parsedStatefulSet.Spec.Template, d.observeConfig)
			if err != nil {
				log.Info("%s", err.Error())
			}

			kubernetesResourceMetaInfo := types.KubernetesAPIObjectMetaInformation{
				APIVersion:   apiVersion,
				ResourceType: statefulSetResourceType,
				Namespace:    parsedStatefulSet.Namespace,
				WorkloadName: parsedStatefulSet.Name,
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
