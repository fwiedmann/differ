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
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/fwiedmann/differ/pkg/observer"

	tags_analyzer "github.com/fwiedmann/differ/pkg/tags-analyzer"

	"github.com/fwiedmann/differ/pkg/image"

	log "github.com/sirupsen/logrus"

	"github.com/fwiedmann/differ/pkg/registries/api"

	"go.uber.org/ratelimit"
)

func StartNewImageWorker(ctx context.Context, rateLimiter ratelimit.Limiter, info chan<- Event) *Worker {
	newWorker := Worker{
		associatedKubernetesObjects: make(map[string]observer.ImageWithKubernetesMetadata),
		mutex:                       sync.RWMutex{},
		informChan:                  info,
		rateLimiter:                 rateLimiter,
	}
	go newWorker.startRunning(ctx)
	return &newWorker
}

type Worker struct {
	associatedKubernetesObjects map[string]observer.ImageWithKubernetesMetadata
	latestTag                   string
	mutex                       sync.RWMutex
	informChan                  chan<- Event
	rateLimiter                 ratelimit.Limiter
	client                      *api.Client
}

func (w *Worker) startRunning(ctx context.Context) {
	imageWorkerCtx, cancel := context.WithCancel(ctx)

imageWorkerRoutine:
	for {
		if len(w.associatedKubernetesObjects) == 0 {
			continue
		}

		if w.client == nil {
			for _, obj := range w.associatedKubernetesObjects {
				w.client = api.New(http.Client{
					Timeout: time.Second * 10,
				}, obj.ImageWithPullSecrets)
				break
			}
		}
		select {
		case <-imageWorkerCtx.Done():
			cancel()
			break imageWorkerRoutine
		default:
			w.mutex.RLock()

			tags, err := w.requestTagsFromAPIWithAllStoredObjects(ctx)
			if err != nil {
				log.Errorf(err.Error())
				w.mutex.RUnlock()
				continue
			}
			w.sendEventForEachStoredObjectIfNewerTagExits(tags)
			w.mutex.RUnlock()
		}
	}
}

func (w *Worker) requestTagsFromAPIWithAllStoredObjects(ctx context.Context) ([]string, error) {
	var imageName string
	var latestError error
	for _, obj := range w.associatedKubernetesObjects {
		imageName = obj.ImageWithPullSecrets.GetNameWithRegistry()
		if len(obj.ImageWithPullSecrets.GetPullSecrets()) == 0 {
			tags, err := w.requestTagsWithRateLimit(ctx, nil)
			if err == nil {
				return tags, nil
			}
			latestError = err
		} else {
			tags, err := w.requestTagsFromAPIWithSecrets(ctx, obj.ImageWithPullSecrets.GetPullSecrets())
			if err == nil {
				return tags, nil
			}
			latestError = err
		}
	}
	return nil, fmt.Errorf("could not fetch any tags for image %s, error: %s", imageName, latestError)
}

func (w *Worker) requestTagsFromAPIWithSecrets(ctx context.Context, secrets []image.PullSecret) ([]string, error) {
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

func (w *Worker) requestTagsWithRateLimit(ctx context.Context, s api.PullSecret) ([]string, error) {
	w.rateLimiter.Take()
	return w.client.GetTagsForImage(ctx, s)
}

func (w *Worker) sendEventForEachStoredObjectIfNewerTagExits(allTagsFromRegistry []string) {
	for _, obj := range w.associatedKubernetesObjects {
		go w.sendEventForStoredObjectIfNewerTagExits(obj, allTagsFromRegistry)
	}
}

func (w *Worker) sendEventForStoredObjectIfNewerTagExits(objectEvent observer.ImageWithKubernetesMetadata, allTagsFromRegistry []string) {
	tagExpr, err := tags_analyzer.GetRegexExprForTag(objectEvent.ImageWithPullSecrets.GetTag())
	if err != nil {
		log.Errorf("could not get a tag expression for image %s with tag %s", objectEvent.ImageWithPullSecrets.GetNameWithRegistry(), objectEvent.ImageWithPullSecrets.GetTag())
		return
	}
	latestTag, err := tags_analyzer.GetLatestTagWithRegexExpr(allTagsFromRegistry, tagExpr)
	if err != nil {
		log.Errorf("registry/worker: %s", err)
	}
	if objectEvent.ImageWithPullSecrets.GetTag() == latestTag {
		return
	}

	w.informChan <- Event{
		ImageWithKubernetesMetadata: objectEvent,
		latestTag:                   latestTag,
	}

}

func (w *Worker) AddOrUpdateAssociatedKubernetesObjects(obj observer.ImageWithKubernetesMetadata) {
	w.mutex.Lock()
	w.associatedKubernetesObjects[obj.GetUID()] = obj
	w.mutex.Unlock()
}

func (w *Worker) DeleteAssociatedKubernetesObjects(obj observer.ImageWithKubernetesMetadata) {
	w.mutex.Lock()
	delete(w.associatedKubernetesObjects, obj.GetUID())
	w.mutex.Unlock()
}
