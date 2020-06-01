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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fwiedmann/differ/pkg/monitoring"

	"github.com/fwiedmann/differ/pkg/registry"

	tagsanalyzer "github.com/fwiedmann/differ/pkg/tags-analyzing"
	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
)

type OciRegistryAPIClient interface {
	GetTagsForImage(ctx context.Context, secret registry.OciPullSecret) ([]string, error)
}

type ListImagesRepository interface {
	ListImages(ctx context.Context, opts ListOptions) ([]Image, error)
}

func StartNewImageWorker(ctx context.Context, client OciRegistryAPIClient, registry, imageName string, rateLimiter ratelimit.Limiter, info chan<- NotificationEvent, repository ListImagesRepository, workerAPIRequestSleepDuration time.Duration) *Worker {
	newWorker := Worker{
		client:                  client,
		registry:                registry,
		imageName:               imageName,
		rp:                      repository,
		mutex:                   sync.RWMutex{},
		informChan:              info,
		rateLimiter:             rateLimiter,
		stop:                    make(chan struct{}),
		apiRequestSleepDuration: workerAPIRequestSleepDuration,
	}
	go newWorker.startRunning(ctx)
	return &newWorker
}

type Worker struct {
	imageName               string
	registry                string
	rp                      ListImagesRepository
	mutex                   sync.RWMutex
	informChan              chan<- NotificationEvent
	rateLimiter             ratelimit.Limiter
	client                  OciRegistryAPIClient
	stop                    chan struct{}
	apiRequestSleepDuration time.Duration
}

func (w *Worker) Stop() {
	w.stop <- struct{}{}
	close(w.stop)
}

func (w *Worker) startRunning(ctx context.Context) {
	imageWorkerCtx, cancel := context.WithCancel(ctx)

	for {
		select {
		case <-imageWorkerCtx.Done():
			cancel()
			return
		case <-w.stop:
			cancel()
			return
		default:
			time.Sleep(w.apiRequestSleepDuration)
			images, err := w.rp.ListImages(imageWorkerCtx, ListOptions{
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

			w.mutex.RLock()

			tags, err := w.requestTagsFromAPIWithAllStoredObjects(ctx, images)
			if err != nil {
				w.updateOCIRegistryMetrics(err)
				log.Warnf(err.Error())
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
	return nil, fmt.Errorf("differentiate/oci-worker error: could not fetch any tags for Image %s, error: %w", imageName, latestError)
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
	return nil, fmt.Errorf("differentiate/oci-worker error: no pull secrets provided to request")
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
	tagExpr, err := tagsanalyzer.GetRegexExprForTag(img.Tag)
	if err != nil {
		log.Errorf("differentiate/oci-worker error: could not get a tag expression for Image %s with tag %s", img.GetNameWithRegistry(), img.Tag)
		return
	}

	latestTag, err := tagsanalyzer.GetLatestTagWithRegexExpr(allTagsFromRegistry, tagExpr)
	if err != nil {
		log.Errorf("differentiate/oci-worker error: %s", err)
	}

	if img.Tag == latestTag {
		return
	}

	monitoring.OciImageNewerTagAvailableMetric.WithLabelValues(img.GetNameWithRegistry(), img.GetRegistryURL(), img.Tag, latestTag, tagExpr.String()).Set(1)
	w.informChan <- NotificationEvent{Image: img, NewTag: latestTag}
}

func (w *Worker) updateOCIRegistryMetrics(err error) {
	if errors.Is(err, registry.StatusUnauthorizedError) {
		monitoring.OciRegistryUnauthorizedErrorMetric.WithLabelValues(w.imageName, w.registry).Inc()
	}
	if errors.Is(err, registry.StatusForbiddenError) {
		monitoring.OciRegistryForbiddenErrorMetric.WithLabelValues(w.imageName, w.registry).Inc()
	}
	if errors.Is(err, registry.StatusToManyRequests) {
		monitoring.OciRegistryToManyRequestsErrorMetric.WithLabelValues(w.imageName, w.registry).Inc()
	}
	if err != nil {
		monitoring.OciRegistryNoTagsFoundMetric.WithLabelValues(w.imageName, w.registry).Inc()
	}
}
