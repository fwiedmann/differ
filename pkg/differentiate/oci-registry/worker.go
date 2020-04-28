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
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/fwiedmann/differ/pkg/differentiate"
	"github.com/fwiedmann/differ/pkg/differentiate/oci-registry/api"

	tags_analyzer "github.com/fwiedmann/differ/pkg/tags-analyzer"
	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
)

func StartNewImageWorker(ctx context.Context, registry, imageName string, rateLimiter ratelimit.Limiter, info chan<- differentiate.NotificationEvent, repository differentiate.Repository) *Worker {

	newWorker := Worker{
		registry:    imageName,
		imageName:   registry,
		rp:          repository,
		mutex:       sync.RWMutex{},
		informChan:  info,
		rateLimiter: rateLimiter,
	}
	go newWorker.startRunning(ctx)
	return &newWorker
}

type Worker struct {
	imageName   string
	registry    string
	rp          differentiate.Repository
	mutex       sync.RWMutex
	informChan  chan<- differentiate.NotificationEvent
	rateLimiter ratelimit.Limiter
	client      *api.Client
}

func (w *Worker) startRunning(ctx context.Context) {
	imageWorkerCtx, cancel := context.WithCancel(ctx)

imageWorkerRoutine:
	for {

		images, err := w.rp.ListImages(imageWorkerCtx, &differentiate.ListOptions{
			ImageName: w.imageName,
			Registry:  w.registry,
		})

		if err != nil {
			log.Error(err)
			continue
		}

		if len(images) == 0 {
			continue
		}

		if w.client == nil {
			w.client = api.New(http.Client{Timeout: time.Second * 10}, images[0])
		}

		select {
		case <-imageWorkerCtx.Done():
			cancel()
			break imageWorkerRoutine
		default:
			w.mutex.RLock()

			tags, err := w.requestTagsFromAPIWithAllStoredObjects(ctx, images)
			if err != nil {
				log.Errorf(err.Error())
				w.mutex.RUnlock()
				continue
			}
			w.sendEventForEachStoredObjectIfNewerTagExits(tags, images)
			w.mutex.RUnlock()
		}
	}
}

func (w *Worker) requestTagsFromAPIWithAllStoredObjects(ctx context.Context, imgs []differentiate.Image) ([]string, error) {
	var imageName string
	var latestError error
	for _, img := range imgs {
		imageName = img.GetNameWithRegistry()
		if len(img.Auth) == 0 {
			tags, err := w.requestTagsWithRateLimit(ctx, nil)
			if err == nil {
				return tags, nil
			}
			latestError = err
			continue
		}

		tags, err := w.requestTagsFromAPIWithSecrets(ctx, img.Auth)
		if err == nil {
			return tags, nil
		}
		latestError = err
	}
	return nil, fmt.Errorf("could not fetch any tags for image %s, error: %s", imageName, latestError)
}

func (w *Worker) requestTagsFromAPIWithSecrets(ctx context.Context, secrets []*differentiate.PullSecret) ([]string, error) {
	for i, secret := range secrets {
		tags, err := w.requestTagsWithRateLimit(ctx, secret)
		if err != nil {
			if i == len(secrets)-1 {
				return nil, err
			}
			continue
		}
		return tags, nil
	}
	return nil, fmt.Errorf("no pull secrets provided to request")
}

func (w *Worker) requestTagsWithRateLimit(ctx context.Context, s *differentiate.PullSecret) ([]string, error) {
	w.rateLimiter.Take()
	return w.client.GetTagsForImage(ctx, s)
}

func (w *Worker) sendEventForEachStoredObjectIfNewerTagExits(allTagsFromRegistry []string, imgs []differentiate.Image) {
	for _, img := range imgs {
		go w.sendEventForStoredObjectIfNewerTagExits(img, allTagsFromRegistry)
	}
}

func (w *Worker) sendEventForStoredObjectIfNewerTagExits(img differentiate.Image, allTagsFromRegistry []string) {
	tagExpr, err := tags_analyzer.GetRegexExprForTag(img.Tag)
	if err != nil {
		log.Errorf("could not get a tag expression for image %s with tag %s", img.GetNameWithRegistry(), img.Tag)
		return
	}

	latestTag, err := tags_analyzer.GetLatestTagWithRegexExpr(allTagsFromRegistry, tagExpr)
	if err != nil {
		log.Errorf("registry/worker: %s", err)
	}

	if img.Tag == latestTag {
		return
	}

	w.informChan <- differentiate.NotificationEvent{Image: img, NewTag: latestTag}
}
