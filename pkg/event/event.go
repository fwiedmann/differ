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

// ObservedKubernetesAPIObjectEvent contains unique meta information from scraped resource types
type ObservedKubernetesAPIObjectEvent struct {
	MetaInformation      KubernetesAPIObjectMetaInformation
	ImageWithPullSecrets image.WithAssociatedPullSecrets
}

// KubernetesEventCommunicationChannels for communication between differ controller and kubernetes api observers
type KubernetesEventCommunicationChannels struct {
	addEventChannel, deleteEventChannel, updateEventChannel chan ObservedKubernetesAPIObjectEvent
	errorEventChannel                                       chan error
}

// NewCommunicationChannels with each channel of the size of the observer count
func NewCommunicationChannels(observerCount int) KubernetesEventCommunicationChannels {
	return KubernetesEventCommunicationChannels{
		addEventChannel:    make(chan ObservedKubernetesAPIObjectEvent, observerCount),
		deleteEventChannel: make(chan ObservedKubernetesAPIObjectEvent, observerCount),
		updateEventChannel: make(chan ObservedKubernetesAPIObjectEvent, observerCount),
		errorEventChannel:  make(chan error, observerCount),
	}
}

// GetADDReceiverEventChanel return the event type ADD channel for the differ controller
func (eventChannels KubernetesEventCommunicationChannels) GetADDReceiverEventChanel() <-chan ObservedKubernetesAPIObjectEvent {
	return eventChannels.addEventChannel
}

// GetUPDATEReceiverEventChanel return the event type UPDATE channel for the differ controller
func (eventChannels KubernetesEventCommunicationChannels) GetUPDATEReceiverEventChanel() <-chan ObservedKubernetesAPIObjectEvent {
	return eventChannels.updateEventChannel
}

// GetDELETReceiverEventChanel return the event type DELETE channel for the differ controller
func (eventChannels KubernetesEventCommunicationChannels) GetDELETReceiverEventChanel() <-chan ObservedKubernetesAPIObjectEvent {
	return eventChannels.deleteEventChannel
}

// GetERRORReceiverEventChanel return the event type ERROR channel for the differ controller
func (eventChannels KubernetesEventCommunicationChannels) GetERRORReceiverEventChanel() <-chan error {
	return eventChannels.errorEventChannel
}

// SendADDEventsToReceiver of event type ADD for all containers of a kubernetes API object
func (eventChannels KubernetesEventCommunicationChannels) SendADDEventsToReceiver(events []ObservedKubernetesAPIObjectEvent) {
	sendEvents(eventChannels.addEventChannel, events)
}

// SendUPDATEEventsToReceiver of event type UPDATE for all containers of a kubernetes API object
func (eventChannels KubernetesEventCommunicationChannels) SendUPDATEEventsToReceiver(events []ObservedKubernetesAPIObjectEvent) {
	sendEvents(eventChannels.updateEventChannel, events)
}

// SendDELETEEventsToReceiver of event type DELETE for all containers of a kubernetes API object
func (eventChannels KubernetesEventCommunicationChannels) SendDELETEEventsToReceiver(events []ObservedKubernetesAPIObjectEvent) {
	sendEvents(eventChannels.deleteEventChannel, events)
}

// SendERRORToReceiver of event type ERROR when an observer occurred an error
func (eventChannels KubernetesEventCommunicationChannels) SendERRORToReceiver(err error) {
	eventChannels.errorEventChannel <- err
}

func sendEvents(receiver chan<- ObservedKubernetesAPIObjectEvent, events []ObservedKubernetesAPIObjectEvent) {
	for _, eventToSend := range events {
		receiver <- eventToSend
	}
}
