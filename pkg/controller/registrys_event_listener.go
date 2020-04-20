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

	"github.com/fwiedmann/differ/pkg/registries/worker"

	log "github.com/sirupsen/logrus"
)

type RegistryEventListener struct {
	eventChan <-chan worker.Event
}

func NewRegistryEventListener(eventChan <-chan worker.Event) RegistryEventListener {
	return RegistryEventListener{eventChan: eventChan}
}

func (r *RegistryEventListener) Start(ctx context.Context) {
	infoCtx, cancel := context.WithCancel(ctx)
differImageEventMonitorRoutine:
	for {
		select {
		case newEvent := <-r.eventChan:
			log.Infof("%+v", newEvent.GeLatestTag())
		case <-infoCtx.Done():
			cancel()
			break differImageEventMonitorRoutine
		}
	}
}
