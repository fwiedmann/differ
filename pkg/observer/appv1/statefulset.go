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

	"k8s.io/client-go/informers"

	"github.com/fwiedmann/differ/pkg/event"

	"github.com/fwiedmann/differ/pkg/observer"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/tools/cache"
)

type StatefulSet struct {
	observerConfig           observer.Config
	kubernetesSharedInformer cache.SharedIndexInformer
}

func NewStatefulSetObserver() *StatefulSet {
	return &StatefulSet{}
}

func (statefulSetObserver *StatefulSet) InitObserverWithKubernetesSharedInformer(observerConfig observer.Config) {
	factory := observer.InitNewKubernetesFactory(observerConfig)
	statefulSetObserver.kubernetesSharedInformer = statefulSetObserver.initStatefulSetSharedInformer(factory)
	statefulSetObserver.observerConfig = observerConfig
}

func (statefulSetObserver *StatefulSet) initStatefulSetSharedInformer(factory informers.SharedInformerFactory) cache.SharedIndexInformer {
	informer := factory.Apps().V1().StatefulSets().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    statefulSetObserver.onAdd,
		DeleteFunc: statefulSetObserver.onDelete,
		UpdateFunc: statefulSetObserver.onUpdate,
	})
	return informer
}

func (statefulSetObserver *StatefulSet) StartObserving() {
	statefulSetObserver.kubernetesSharedInformer.Run(statefulSetObserver.observerConfig.StopEventCommunicationChannel)
}

func (statefulSetObserver *StatefulSet) onAdd(obj interface{}) {
	statefulSetObserver.sendObjectToEventReceiverType(obj, statefulSetObserver.observerConfig.SendADDEventsToReceiver)
}

func (statefulSetObserver *StatefulSet) onUpdate(_, newObj interface{}) {
	statefulSetObserver.sendObjectToEventReceiverType(newObj, statefulSetObserver.observerConfig.SendUPDATEEventsToReceiver)
}

func (statefulSetObserver *StatefulSet) onDelete(obj interface{}) {
	statefulSetObserver.sendObjectToEventReceiverType(obj, statefulSetObserver.observerConfig.SendDELETEEventsToReceiver)
}

func (statefulSetObserver *StatefulSet) sendObjectToEventReceiverType(obj interface{}, sender func(events []event.ObservedKubernetesAPIObjectEvent)) {
	statefulSet, err := statefulSetObserver.convertToResource(obj)
	if err != nil {
		statefulSetObserver.observerConfig.SendERRORToReceiver(err)
	} else {
		kubernetesResourceMetaInfo := event.NewKubernetesAPIObjectMetaInformation(apiVersion, statefulSteSetResourceType, statefulSet.Namespace, statefulSet.Name)
		eventsToSend, err := statefulSetObserver.observerConfig.EventGenerator.GenerateEventsFromPodSpec(statefulSet.Spec.Template.Spec, kubernetesResourceMetaInfo)
		if err != nil {
			statefulSetObserver.observerConfig.SendERRORToReceiver(err)
		} else {
			sender(eventsToSend)
		}
	}
}

func (statefulSetObserver *StatefulSet) convertToResource(obj interface{}) (*v1.StatefulSet, error) {
	staefulSet, ok := obj.(*v1.StatefulSet)
	if !ok {
		return nil, errors.New("could not parse statefulSetObserver object")
	}
	return staefulSet, nil
}
