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
	"fmt"
	"sync"

	tags_analyzer "github.com/fwiedmann/differ/pkg/tags-analyzer"
	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
)

type OciRegistryAPIClient interface {
	GetTagsForImage(ctx context.Context, secret OciPullSecret) ([]string, error)
}

func StartNewImageWorker(ctx context.Context, client OciRegistryAPIClient, registry, imageName string, rateLimiter ratelimit.Limiter, info chan<- NotificationEvent, repository Repository) *Worker {

	newWorker := Worker{
		client:      client,
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
	rp          Repository
	mutex       sync.RWMutex
	informChan  chan<- NotificationEvent
	rateLimiter ratelimit.Limiter
	client      OciRegistryAPIClient
}

func (w *Worker) startRunning(ctx context.Context) {
	imageWorkerCtx, cancel := context.WithCancel(ctx)

imageWorkerRoutine:
	for {

		images, err := w.rp.ListImages(imageWorkerCtx, &ListOptions{
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

func (w *Worker) requestTagsFromAPIWithAllStoredObjects(ctx context.Context, imgs []Image) ([]string, error) {
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

func (w *Worker) requestTagsFromAPIWithSecrets(ctx context.Context, secrets []*PullSecret) ([]string, error) {
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

func (w *Worker) requestTagsWithRateLimit(ctx context.Context, s *PullSecret) ([]string, error) {
	w.rateLimiter.Take()
	return w.client.GetTagsForImage(ctx, s)
}

func (w *Worker) sendEventForEachStoredObjectIfNewerTagExits(allTagsFromRegistry []string, imgs []Image) {
	for _, img := range imgs {
		go w.sendEventForStoredObjectIfNewerTagExits(img, allTagsFromRegistry)
	}
}

func (w *Worker) sendEventForStoredObjectIfNewerTagExits(img Image, allTagsFromRegistry []string) {
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

	w.informChan <- NotificationEvent{Image: img, NewTag: latestTag}
}
