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

package event

import (
	"github.com/fwiedmann/differ/pkg/image"
)

type KubernetesAPIObjectMetaInformation struct {
	APIVersion   string
	ResourceType string
	Namespace    string
	WorkloadName string
}

// ObservedKubernetesAPIObjectEvent contains unique meta information from scraped resource types
type ObservedKubernetesAPIObjectEvent struct {
	MetaInformation      KubernetesAPIObjectMetaInformation
	ImageWithPullSecrets image.WithAssociatedPullSecrets
}

type KubernetesEventCommunicationChannels struct {
	addEventChannel, deleteEventChannel, updateEventChannel chan ObservedKubernetesAPIObjectEvent
	StopEventCommunicationChannel                           chan struct{}
	errorEventChannel                                       chan error
}

func InitCommunicationChannels() KubernetesEventCommunicationChannels {
	return KubernetesEventCommunicationChannels{
		addEventChannel:               make(chan ObservedKubernetesAPIObjectEvent),
		deleteEventChannel:            make(chan ObservedKubernetesAPIObjectEvent),
		updateEventChannel:            make(chan ObservedKubernetesAPIObjectEvent),
		StopEventCommunicationChannel: make(chan struct{}),
		errorEventChannel:             make(chan error),
	}
}
func (eventChannels KubernetesEventCommunicationChannels) GetADDReceiverEventChanel() <-chan ObservedKubernetesAPIObjectEvent {
	return eventChannels.addEventChannel
}
func (eventChannels KubernetesEventCommunicationChannels) GetUPDATEReceiverEventChanel() <-chan ObservedKubernetesAPIObjectEvent {
	return eventChannels.updateEventChannel
}

func (eventChannels KubernetesEventCommunicationChannels) GetDELETReceiverEventChanel() <-chan ObservedKubernetesAPIObjectEvent {
	return eventChannels.deleteEventChannel
}

func (eventChannels KubernetesEventCommunicationChannels) GetERRORReceiverEventChanel() <-chan error {
	return eventChannels.errorEventChannel
}

func (eventChannels KubernetesEventCommunicationChannels) SendADDEventsToReceiver(events []ObservedKubernetesAPIObjectEvent) {
	sendEvents(eventChannels.addEventChannel, events)
}

func (eventChannels KubernetesEventCommunicationChannels) SendUPDATEEventsToReceiver(events []ObservedKubernetesAPIObjectEvent) {
	sendEvents(eventChannels.updateEventChannel, events)
}

func (eventChannels KubernetesEventCommunicationChannels) SendDELETEEventsToReceiver(events []ObservedKubernetesAPIObjectEvent) {
	sendEvents(eventChannels.deleteEventChannel, events)
}

func (eventChannels KubernetesEventCommunicationChannels) SendERRORToReceiver(err error) {
	eventChannels.errorEventChannel <- err
}

func sendEvents(receiver chan<- ObservedKubernetesAPIObjectEvent, events []ObservedKubernetesAPIObjectEvent) {
	for _, eventToSend := range events {
		receiver <- eventToSend
	}
}

func NewKubernetesAPIObjectMetaInformation(apiVersion, observedAPIResource, namespace, resourceName string) KubernetesAPIObjectMetaInformation {
	return KubernetesAPIObjectMetaInformation{
		APIVersion:   apiVersion,
		ResourceType: observedAPIResource,
		Namespace:    namespace,
		WorkloadName: resourceName,
	}
}
