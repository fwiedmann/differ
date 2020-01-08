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

package appv1

import (
	"errors"

	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/fwiedmann/differ/pkg/event"
	"github.com/fwiedmann/differ/pkg/observer"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

type DaemonSet struct {
	observerConfig           observer.Config
	kubernetesSharedInformer cache.SharedIndexInformer
	stopChannel              chan struct{}
}

func NewDaemonSetObserver() *DaemonSet {
	return &DaemonSet{}
}

func (daemonSetObserver *DaemonSet) InitObserverWithKubernetesSharedInformer(observerConfig observer.Config) {
	factory := observer.InitNewKubernetesFactory(observerConfig)
	daemonSetObserver.kubernetesSharedInformer = daemonSetObserver.initDaemonSettSharedInformer(factory)
	daemonSetObserver.observerConfig = observerConfig
	daemonSetObserver.stopChannel = make(chan struct{})
}

func (daemonSetObserver *DaemonSet) initDaemonSettSharedInformer(factory informers.SharedInformerFactory) cache.SharedIndexInformer {
	informer := factory.Apps().V1().DaemonSets().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    daemonSetObserver.onAdd,
		DeleteFunc: daemonSetObserver.onDelete,
		UpdateFunc: daemonSetObserver.onUpdate,
	})
	return informer
}

func (daemonSetObserver *DaemonSet) StartObserving() {
	daemonSetObserver.kubernetesSharedInformer.Run(daemonSetObserver.stopChannel)
}

func (daemonSetObserver *DaemonSet) StopObserving() {
	defer runtime.HandleCrash()
	daemonSetObserver.stopChannel <- struct{}{}
	close(daemonSetObserver.stopChannel)
}

func (daemonSetObserver *DaemonSet) onAdd(obj interface{}) {
	daemonSetObserver.sendObjectToEventReceiverType(obj, daemonSetObserver.observerConfig.SendADDEventsToReceiver)
}

func (daemonSetObserver *DaemonSet) onUpdate(_, newObj interface{}) {
	daemonSetObserver.sendObjectToEventReceiverType(newObj, daemonSetObserver.observerConfig.SendUPDATEEventsToReceiver)
}

func (daemonSetObserver *DaemonSet) onDelete(obj interface{}) {
	daemonSetObserver.sendObjectToEventReceiverType(obj, daemonSetObserver.observerConfig.SendDELETEEventsToReceiver)
}

func (daemonSetObserver *DaemonSet) sendObjectToEventReceiverType(obj interface{}, sender func(events []event.ObservedKubernetesAPIObjectEvent)) {
	daemonSet, err := daemonSetObserver.convertToResource(obj)
	if err != nil {
		daemonSetObserver.observerConfig.SendERRORToReceiver(err)
	} else {
		kubernetesResourceMetaInfo := event.NewKubernetesAPIObjectMetaInformation(daemonSet.GetUID(), apiVersion, daemonSetResourceType, daemonSet.Namespace, daemonSet.Name)
		eventsToSend, err := daemonSetObserver.observerConfig.EventGenerator.GenerateEventsFromPodSpec(daemonSet.Spec.Template.Spec, kubernetesResourceMetaInfo)
		if err != nil {
			daemonSetObserver.observerConfig.SendERRORToReceiver(err)
		} else {
			sender(eventsToSend)
		}
	}
}

func (daemonSetObserver *DaemonSet) convertToResource(obj interface{}) (*v1.DaemonSet, error) {
	daemonSet, ok := obj.(*v1.DaemonSet)
	if !ok {
		return nil, errors.New("could not parse DaemonSet object")
	}
	return daemonSet, nil
}
