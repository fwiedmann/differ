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

	"github.com/fwiedmann/differ/pkg/registries/worker"

	log "github.com/sirupsen/logrus"
)

type Store struct {
	instances    *sync.Map
	informerChan chan<- worker.Event
}

func NewRegistriesStore(informerChan chan<- worker.Event) *Store {
	return &Store{
		instances:    &sync.Map{},
		informerChan: informerChan,
	}
}

func (s *Store) AddImage(ctx context.Context, obj observer.ImageWithKubernetesMetadata) {
	registryURL := obj.ImageWithPullSecrets.GetRegistryURL()
	if s.registryIsNotInStoreYet(registryURL) {
		s.instances.Store(registryURL, newRegistry(s.informerChan))
	}
	correspondingRegistryOfImage, found := s.instances.Load(registryURL)
	if !found {
		log.Errorf("tried to add image for object with uuid \"%s\" but corresponding registries \"%s\" does not exists", obj.GetUID(), obj.ImageWithPullSecrets.GetRegistryURL())
		return
	}
	r := correspondingRegistryOfImage.(*registry)
	r.addOrUpdateImage(ctx, obj)
}

func (s *Store) registryIsNotInStoreYet(registryURL string) bool {
	if _, found := s.instances.Load(registryURL); !found {
		return true
	}
	return false
}

func (s *Store) UpdateImage(ctx context.Context, obj observer.ImageWithKubernetesMetadata) {
	correspondingRegistryOfImage, found := s.instances.Load(obj.ImageWithPullSecrets.GetRegistryURL())
	if !found {
		log.Errorf("tried to update image for object with uuid \"%s\" but corresponding registries \"%s\" does not exists", obj.GetUID(), obj.ImageWithPullSecrets.GetRegistryURL())
		return
	}
	r := correspondingRegistryOfImage.(*registry)
	r.addOrUpdateImage(ctx, obj)
}

func (s *Store) DeleteImage(obj observer.ImageWithKubernetesMetadata) {
	correspondingRegistryOfImage, found := s.instances.Load(obj.ImageWithPullSecrets.GetRegistryURL())
	if !found {
		log.Errorf("tried to delete image for object with uuid \"%s\" but corresponding registries \"%s\" does not exists", obj.GetUID(), obj.ImageWithPullSecrets.GetRegistryURL())
		return
	}
	r := correspondingRegistryOfImage.(*registry)
	r.deleteImage(obj)
}
