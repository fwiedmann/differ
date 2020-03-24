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
	"context"

	"github.com/fwiedmann/differ/pkg/registries"

	log "github.com/sirupsen/logrus"

	"github.com/fwiedmann/differ/pkg/event"
)

type Observer interface {
	StartObserving(ctx context.Context)
}

// KubernetesAPIEventListener types struct
type KubernetesAPIEventListener struct {
	kubernetesEventChannels event.KubernetesEventCommunicationChannels
	observers               []Observer
	errChan                 chan<- error
	registriesStore         registries.Store
}

// NewKubernetesAPIEventListener initialize a KubernetesAPIEventListener
func NewKubernetesAPIEventListener(kubernetesEventChannels event.KubernetesEventCommunicationChannels, errorChan chan<- error, observers []Observer, rs registries.Store) KubernetesAPIEventListener {
	return KubernetesAPIEventListener{
		kubernetesEventChannels: kubernetesEventChannels,
		observers:               observers,
		errChan:                 errorChan,
		registriesStore:         rs,
	}
}

// Start starts kubernetes API listener controller
func (k KubernetesAPIEventListener) Start(ctx context.Context) {
	k.startAllObservers(ctx)
	eventCtx, cancel := context.WithCancel(ctx)
	defer cancel()
differEventMonitorRoutine:
	for {
		select {
		case createEvent := <-k.kubernetesEventChannels.GetADDReceiverEventChanel():
			log.Infof("create event: %s", createEvent)
			k.registriesStore.AddImage(eventCtx, createEvent)
		case deleteEvent := <-k.kubernetesEventChannels.GetDELETReceiverEventChanel():
			log.Infof("delete event: %s", deleteEvent)
			k.registriesStore.DeleteImage(deleteEvent)
		case updateEvent := <-k.kubernetesEventChannels.GetUPDATEReceiverEventChanel():
			log.Infof("update event: %s", updateEvent)
			k.registriesStore.UpdateImage(eventCtx, updateEvent)
		case errorEvent := <-k.kubernetesEventChannels.GetERRORReceiverEventChanel():
			log.Errorf("%s", errorEvent)
			break differEventMonitorRoutine
		case <-eventCtx.Done():
			break differEventMonitorRoutine
		}
	}
}

func (k KubernetesAPIEventListener) startAllObservers(ctx context.Context) {
	for _, o := range k.observers {
		go o.StartObserving(ctx)
	}
}
