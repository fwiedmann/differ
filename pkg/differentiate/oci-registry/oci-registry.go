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

package oci_registry

import (
	"context"

	"github.com/fwiedmann/differ/pkg/differentiate"
)

func NewOCIRegistryService(ctx context.Context, rp differentiate.Repository) differentiate.Service {
	ors := &OCIRegistryService{
		rp:                 rp,
		workerNotification: make(chan differentiate.NotificationEvent, 100),
	}
	go ors.multiplexToNotifiers(ctx)

	return ors
}

type OCIRegistryService struct {
	rp                 differentiate.Repository
	workers            map[string]*Worker
	notifiers          []chan<- differentiate.NotificationEvent
	workerNotification chan differentiate.NotificationEvent
}

func (O *OCIRegistryService) AddImage(ctx context.Context, image *differentiate.Image) error {

	if err := O.rp.AddImage(ctx, image); err != nil {
		return err
	}

	var w *Worker
	w, found := O.workers[image.Registry]
	if !found {
		w = StartNewImageWorker(ctx, image.Registry, image.Name, createRateLimitForRegistry(image.Registry), O.workerNotification, O.rp)
		O.workers[image.Registry] = w
	}
	return nil
}

func (O *OCIRegistryService) DeleteImage(ctx context.Context, image *differentiate.Image) error {
	return O.rp.DeleteImage(ctx, image)
}

func (O *OCIRegistryService) UpdateImage(ctx context.Context, image *differentiate.Image) error {
	return O.rp.UpdateImage(ctx, image)
}

func (O *OCIRegistryService) ListImage(ctx context.Context, opts *differentiate.ListOptions) ([]differentiate.Image, error) {
	return O.rp.ListImages(ctx, opts)
}

func (O *OCIRegistryService) Notify(event chan<- differentiate.NotificationEvent) {
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

func closeChannels(c []chan<- differentiate.NotificationEvent) {
	for _, ch := range c {
		close(ch)
	}
}
