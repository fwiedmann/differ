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

type Deployment struct {
	observerConfig           observer.Config
	kubernetesSharedInformer cache.SharedIndexInformer
}

func NewDeploymentObserver() *Deployment {
	return &Deployment{}
}

func (deploymentObserver *Deployment) InitObserverWithKubernetesSharedInformer(observerConfig observer.Config) {
	factory := observer.InitNewKubernetesFactory(observerConfig)
	deploymentObserver.kubernetesSharedInformer = deploymentObserver.initDeploymentSharedInformer(factory)
	deploymentObserver.observerConfig = observerConfig
}

func (deploymentObserver *Deployment) initDeploymentSharedInformer(factory informers.SharedInformerFactory) cache.SharedIndexInformer {
	informer := factory.Apps().V1().Deployments().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    deploymentObserver.onAdd,
		DeleteFunc: deploymentObserver.onDelete,
		UpdateFunc: deploymentObserver.onUpdate,
	})
	return informer
}

func (deploymentObserver *Deployment) StartObserving() {
	deploymentObserver.kubernetesSharedInformer.Run(deploymentObserver.observerConfig.StopEventCommunicationChannel)
}

func (deploymentObserver *Deployment) onAdd(obj interface{}) {
	deploymentObserver.sendObjectToEventReceiverType(obj, deploymentObserver.observerConfig.SendADDEventsToReceiver)
}

func (deploymentObserver *Deployment) onUpdate(_, newObj interface{}) {
	deploymentObserver.sendObjectToEventReceiverType(newObj, deploymentObserver.observerConfig.SendUPDATEEventsToReceiver)
}

func (deploymentObserver *Deployment) onDelete(obj interface{}) {
	deploymentObserver.sendObjectToEventReceiverType(obj, deploymentObserver.observerConfig.SendDELETEEventsToReceiver)
}

func (deploymentObserver *Deployment) sendObjectToEventReceiverType(obj interface{}, sender func(events []event.ObservedKubernetesAPIObjectEvent)) {
	deployment, err := deploymentObserver.convertToResource(obj)
	if err != nil {
		deploymentObserver.observerConfig.SendERRORToReceiver(err)
	} else {
		kubernetesResourceMetaInfo := event.NewKubernetesAPIObjectMetaInformation(apiVersion, deploymentResourceType, deployment.Namespace, deployment.Name)
		eventsToSend, err := deploymentObserver.observerConfig.EventGenerator.GenerateEventsFromPodSpec(deployment.Spec.Template.Spec, kubernetesResourceMetaInfo)
		if err != nil {
			deploymentObserver.observerConfig.SendERRORToReceiver(err)
		} else {
			sender(eventsToSend)
		}
	}
}

func (deploymentObserver *Deployment) convertToResource(obj interface{}) (*v1.Deployment, error) {
	deployment, ok := obj.(*v1.Deployment)
	if !ok {
		return nil, errors.New("could not parse deploymentObserver object")
	}
	return deployment, nil
}
