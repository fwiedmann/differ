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

package worker

import (
	"context"
	"sync"

	"go.uber.org/ratelimit"

	"github.com/fwiedmann/differ/pkg/event"
)

func StartNewImageWorker(ctx context.Context, rateLimiter ratelimit.Limiter, info chan<- event.Tag) *ImageWorker {
	newWorker := ImageWorker{
		associatedKubernetesObjects: make(map[string]event.ObservedKubernetesAPIObjectEvent, 0),
		mutex:                       sync.RWMutex{},
		informChan:                  info,
		rateLimiter:                 rateLimiter,
	}
	go newWorker.startRunning(ctx)
	return &newWorker
}

type ImageWorker struct {
	associatedKubernetesObjects map[string]event.ObservedKubernetesAPIObjectEvent
	latestTag                   string
	mutex                       sync.RWMutex
	informChan                  chan<- event.Tag
	rateLimiter                 ratelimit.Limiter
}

func (iw *ImageWorker) startRunning(ctx context.Context) {
	imageWorkerCtx, cancel := context.WithCancel(ctx)
imageWorkerRoutine:
	for {
		select {
		case <-imageWorkerCtx.Done():
			cancel()
			break imageWorkerRoutine
		default:
			iw.rateLimiter.Take()
			iw.mutex.RLock()
			iw.informChan <- event.NewTag("latest", []event.ObservedKubernetesAPIObjectEvent{})
			iw.mutex.RUnlock()
		}
	}
}

func (iw *ImageWorker) AddOrUpdateAssociatedKubernetesObjects(obj event.ObservedKubernetesAPIObjectEvent) {
	iw.mutex.Lock()
	iw.associatedKubernetesObjects[obj.GetUID()] = obj
	iw.mutex.Unlock()
}

func (iw *ImageWorker) DeleteAssociatedKubernetesObjects(obj event.ObservedKubernetesAPIObjectEvent) {
	iw.mutex.Lock()
	delete(iw.associatedKubernetesObjects, obj.GetUID())
	iw.mutex.Unlock()
}
