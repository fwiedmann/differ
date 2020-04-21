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

package observer

import (
	"context"

	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// EventGenerator for events send from a kubernetes shared index informer
type EventGenerator interface {
	GenerateEventsFromPodSpec(podSpec v1.PodSpec, kubernetesMetaInformation KubernetesAPIObjectMetaInformation) ([]ImageWithKubernetesMetadata, error)
}

// RegistriesStore handle concurrent events observed by the kubernetes shared informer
type RegistriesStore interface {
	AddImage(ctx context.Context, obj ImageWithKubernetesMetadata)
	UpdateImage(ctx context.Context, obj ImageWithKubernetesMetadata)
	DeleteImage(obj ImageWithKubernetesMetadata)
}

// Observer listens on a kubernetes api type events
type Observer struct {
	newKubernetesObjectHandler func(obj interface{}) (KubernetesObjectHandler, error)
	observerConfig             Config
	kubernetesSharedInformer   cache.SharedIndexInformer
	kubernetesObjectKind       string
	kubernetesAPIVersion       string
	ctx                        context.Context
}

// KubernetesObjectHandler extract required information from the kubernetes API type for the event
type KubernetesObjectHandler interface {
	GetPodSpec() v1.PodSpec
	GetNameOfObservedObject() string
	GetUID() types.UID
}

// Config for an observer which is required for the communication
type Config struct {
	namespaceToScrape   string
	kubernetesAPIClient kubernetes.Interface
	eventGenerator      *Generator
	registryStore       RegistriesStore
}

// NewObserverConfig contains a core configuration for each observer
func NewObserverConfig(namespaceToScrape string, kubernetesAPIClient kubernetes.Interface, eventGenerator *Generator, rs RegistriesStore) Config {
	return Config{
		namespaceToScrape:   namespaceToScrape,
		kubernetesAPIClient: kubernetesAPIClient,
		eventGenerator:      eventGenerator,
		registryStore:       rs,
	}

}

// StartObserving of the kubernetes API type and send events to the event channels
func (o *Observer) StartObserving(ctx context.Context) {
	o.ctx = ctx
	observerCtx, cancel := context.WithCancel(ctx)
	stopChan := make(chan struct{})

	go o.kubernetesSharedInformer.Run(stopChan)

	<-observerCtx.Done()
	stopChan <- struct{}{}
	close(stopChan)
	cancel()
}

func (o *Observer) initSharedIndexInformerWithHandleFunctions() {
	o.kubernetesSharedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.onAdd,
		DeleteFunc: o.onDelete,
		UpdateFunc: o.onUpdate,
	})
}

func (o *Observer) onAdd(obj interface{}) {
	observedAPIObjects, err := o.getKubernetesObjects(obj)
	if err != nil {
		log.Error(err)
	}

	for _, obj := range observedAPIObjects {
		go o.observerConfig.registryStore.AddImage(o.ctx, obj)
	}
}

func (o *Observer) onUpdate(_, newObj interface{}) {
	observedAPIObjects, err := o.getKubernetesObjects(newObj)
	if err != nil {
		log.Error(err)
	}

	for _, obj := range observedAPIObjects {
		go o.observerConfig.registryStore.UpdateImage(o.ctx, obj)
	}
}

func (o *Observer) onDelete(obj interface{}) {
	observedAPIObjects, err := o.getKubernetesObjects(obj)
	if err != nil {
		log.Error(err)
	}

	for _, obj := range observedAPIObjects {
		go o.observerConfig.registryStore.DeleteImage(obj)
	}
}

func (o *Observer) getKubernetesObjects(obj interface{}) ([]ImageWithKubernetesMetadata, error) {
	handledObject, err := o.newKubernetesObjectHandler(obj)
	if err != nil {
		return nil, err
	}
	kubernetesResourceMetaInfo := NewKubernetesAPIObjectMetaInformation(handledObject.GetUID(), o.kubernetesAPIVersion, o.kubernetesObjectKind, o.observerConfig.namespaceToScrape, handledObject.GetNameOfObservedObject())
	return o.observerConfig.eventGenerator.GenerateEventsFromPodSpec(handledObject.GetPodSpec(), kubernetesResourceMetaInfo)
}
