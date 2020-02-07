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
	log "github.com/sirupsen/logrus"

	"github.com/fwiedmann/differ/pkg/event"
)

type Observer interface {
	StartObserving()
	StopObserving()
}

// DifferController types struct
type DifferController struct {
	kubernetesEventChannels event.KubernetesEventCommunicationChannels
	observers               []Observer
	errChan                 chan<- error
}

// NewDifferController initialize the differ controller
func NewDifferController(kubernetesEventChannels event.KubernetesEventCommunicationChannels, errorChan chan<- error, observers []Observer) *DifferController {
	return &DifferController{
		kubernetesEventChannels: kubernetesEventChannels,
		observers:               observers,
		errChan:                 errorChan,
	}
}

// StartController starts differ controller loop
func (c *DifferController) StartController() {
	c.startAllObservers()
differMonitorRoutine:
	for {
		select {
		case createEvent := <-c.kubernetesEventChannels.GetADDReceiverEventChanel():
			log.Infof("create event: %+v", createEvent.ImageWithPullSecrets)
		case deleteEvent := <-c.kubernetesEventChannels.GetDELETReceiverEventChanel():
			log.Infof("delete event: %+v", deleteEvent)
		case updateEvent := <-c.kubernetesEventChannels.GetUPDATEReceiverEventChanel():
			log.Infof("update event: %+v", updateEvent)
		case errorEvent := <-c.kubernetesEventChannels.GetERRORReceiverEventChanel():
			c.handleError(errorEvent)
			break differMonitorRoutine
		}
	}
}

func (c *DifferController) handleError(err error) {
	log.Errorf("Gracefully shutdown controller...")
	c.StopAllObservers()
	c.errChan <- err
}

func (c *DifferController) startAllObservers() {
	for _, o := range c.observers {
		go o.StartObserving()
	}
}

func (c *DifferController) StopAllObservers() {
	for _, observerToStop := range c.observers {
		observerToStop.StopObserving()
	}
}
