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

package differentiate

import (
	"context"
	"net/http"
	"sync"
	"time"

	"go.uber.org/ratelimit"
)

type OCIWorker interface {
	Stop()
}

func NewOCIRegistryService(ctx context.Context, rp Repository, startWorker func(ctx context.Context, client OciRegistryAPIClient, registry, imageName string, rateLimiter ratelimit.Limiter, info chan<- NotificationEvent, repository ListImagesRepository) OCIWorker) Service {
	ors := &OCIRegistryService{
		rp:                 rp,
		workerNotification: make(chan NotificationEvent, 100),
		startWorkerFun:     startWorker,
	}
	go ors.multiplexToNotifiers(ctx)

	return ors
}

type OCIRegistryService struct {
	rp                 Repository
	workers            map[string]OCIWorker
	notifiers          []chan<- NotificationEvent
	workerNotification chan NotificationEvent
	workerMtx          sync.Mutex
	startWorkerFun     func(ctx context.Context, client OciRegistryAPIClient, registry, imageName string, rateLimiter ratelimit.Limiter, info chan<- NotificationEvent, repository ListImagesRepository) OCIWorker
}

func (O *OCIRegistryService) AddImage(ctx context.Context, image Image) error {

	if err := O.rp.AddImage(ctx, image); err != nil {
		return err
	}

	_, found := O.workers[image.GetNameWithRegistry()]
	if !found {
		O.workerMtx.Lock()
		O.workers[image.GetNameWithRegistry()] = O.startWorkerFun(ctx, NewOCIAPIClient(http.Client{Timeout: time.Second * 10}, image), image.Registry, image.Name, createRateLimitForRegistry(image.Registry), O.workerNotification, O.rp)
		O.workerMtx.Unlock()
	}
	return nil
}

func (O *OCIRegistryService) DeleteImage(ctx context.Context, image Image) error {
	O.workerMtx.Lock()
	images, err := O.rp.ListImages(ctx, ListOptions{ImageName: image.Name, Registry: image.Registry})
	if err != nil {
		O.workerMtx.Unlock()
		return err
	}
	if len(images) <= 1 {
		if worker, ok := O.workers[image.GetNameWithRegistry()]; ok {
			worker.Stop()
			delete(O.workers, image.GetNameWithRegistry())
		}
	}
	O.workerMtx.Unlock()

	return O.rp.DeleteImage(ctx, image)
}

func (O *OCIRegistryService) UpdateImage(ctx context.Context, image Image) error {
	return O.rp.UpdateImage(ctx, image)
}

func (O *OCIRegistryService) ListImages(ctx context.Context, opts ListOptions) ([]Image, error) {
	return O.rp.ListImages(ctx, opts)
}

func (O *OCIRegistryService) Notify(event chan<- NotificationEvent) {
	O.notifiers = append(O.notifiers, event)
}

func (O *OCIRegistryService) multiplexToNotifiers(ctx context.Context) {
	notifyContext, cancel := context.WithCancel(ctx)
	for {
		select {
		case event := <-O.workerNotification:
			for _, ch := range O.notifiers {
				ch <- event
			}
		case <-notifyContext.Done():
			cancel()
			closeChannels(O.notifiers)
			return
		}
	}
}

func closeChannels(c []chan<- NotificationEvent) {
	for _, ch := range c {
		close(ch)
	}
}
