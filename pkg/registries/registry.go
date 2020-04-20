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

package registries

import (
	"context"
	"sync"

	"github.com/fwiedmann/differ/pkg/observer"

	"go.uber.org/ratelimit"

	"github.com/fwiedmann/differ/pkg/registries/worker"

	log "github.com/sirupsen/logrus"
)

func newRegistry(informChan chan<- worker.Event) *registry {
	return &registry{
		imageWorkers: make(map[string]*worker.Worker),
		mutex:        sync.RWMutex{},
		informChan:   informChan,
		rateLimiter:  ratelimit.New(5),
	}
}

type registry struct {
	imageWorkers map[string]*worker.Worker
	mutex        sync.RWMutex
	informChan   chan<- worker.Event
	rateLimiter  ratelimit.Limiter
}

func (r *registry) addOrUpdateImage(ctx context.Context, obj observer.ImageWithKubernetesMetadata) {
	imageName := obj.ImageWithPullSecrets.GetNameWithRegistry()
	if r.imageIsNotStoredYet(imageName) {
		r.createNewImageWorkerEntry(ctx, imageName)
	}
	r.mutex.RLock()
	correspondingImageWorkerForObject, found := r.imageWorkers[imageName]
	r.mutex.RUnlock()
	if !found {
		log.Errorf("tried to add/update image worker for object with uuid \"%s\" but image worker does not exists", obj.GetUID())
		return
	}
	correspondingImageWorkerForObject.AddOrUpdateAssociatedKubernetesObjects(obj)
}

func (r *registry) imageIsNotStoredYet(imageName string) bool {
	r.mutex.RLock()
	_, found := r.imageWorkers[imageName]
	r.mutex.RUnlock()
	return !found
}

func (r *registry) createNewImageWorkerEntry(ctx context.Context, imageName string) {
	r.mutex.Lock()
	r.imageWorkers[imageName] = worker.StartNewImageWorker(ctx, r.rateLimiter, r.informChan)
	r.mutex.Unlock()
}

func (r *registry) deleteImage(obj observer.ImageWithKubernetesMetadata) {
	imageName := obj.ImageWithPullSecrets.GetNameWithRegistry()
	r.mutex.RLock()
	correspondingImageWorkerForObject, found := r.imageWorkers[imageName]
	r.mutex.RUnlock()
	if !found {
		log.Errorf("tried to delete object with uuid \"%s\" but corresponding image worker does not exists", obj.GetUID())
		return
	}
	correspondingImageWorkerForObject.DeleteAssociatedKubernetesObjects(obj)
}
