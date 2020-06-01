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

package differentiating

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/fwiedmann/differ/pkg/registry"
)

func NewOCIRegistryService(ctx context.Context, rp Repository, workerAPIRequestSleepDuration time.Duration, initOCIAPIClientFun func(c http.Client, img registry.OciImage) OciRegistryAPIClient) Service {
	ors := &OCIRegistryService{
		rp:                              rp,
		registryAPIRequestSleepDuration: workerAPIRequestSleepDuration,
		initOCIAPIClientFun:             initOCIAPIClientFun,
		workerCtx:                       ctx,
		workerNotification:              make(chan NotificationEvent, 100),
		workers:                         make(map[string]*Worker),
	}
	go ors.multiplexToNotifiers(ctx)
	return ors
}

type OCIRegistryService struct {
	registryAPIRequestSleepDuration time.Duration
	rp                              Repository
	workers                         map[string]*Worker
	notifiers                       []chan<- NotificationEvent
	workerNotification              chan NotificationEvent
	workerMtx                       sync.Mutex
	workerCtx                       context.Context
	initOCIAPIClientFun             func(c http.Client, img registry.OciImage) OciRegistryAPIClient
}

func (O *OCIRegistryService) AddImage(ctx context.Context, image Image) error {

	if err := O.rp.AddImage(ctx, image); err != nil {
		return err
	}

	_, found := O.workers[image.GetNameWithRegistry()]
	if !found {
		O.workerMtx.Lock()
		httpClient := http.Client{
			Timeout: time.Second * 10,
		}
		O.workers[image.GetNameWithRegistry()] = StartNewImageWorker(O.workerCtx, O.initOCIAPIClientFun(httpClient, image), image.Registry, image.Name, createRateLimitForRegistry(image.Registry), O.workerNotification, O.rp, O.registryAPIRequestSleepDuration)
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
				go func(ch chan<- NotificationEvent) {
					ch <- event
				}(ch)
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
