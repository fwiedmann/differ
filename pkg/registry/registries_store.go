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

package registry

import (
	"context"
	"sync"

	"github.com/fwiedmann/differ/pkg/event"
	log "github.com/sirupsen/logrus"
)

type Registries struct {
	store        *sync.Map
	informerChan chan<- struct{}
}

func NewRegistriesStore(informerChan chan<- struct{}) Registries {
	return Registries{
		store:        &sync.Map{},
		informerChan: informerChan,
	}
}

func (rs Registries) AddImage(ctx context.Context, obj event.ObservedKubernetesAPIObjectEvent) {
	registryURL := obj.ImageWithPullSecrets.GetRegistryURL()
	if rs.registryIsNotInStoreYet(registryURL) {
		rs.store.Store(registryURL, newRegistry(rs.informerChan))
	}
	correspondingRegistryOfImage, found := rs.store.Load(registryURL)
	if !found {
		log.Errorf("tried to add image for object with uuid \"%s\" but corresponding registry \"%s\" does not exists", obj.GetUID(), obj.ImageWithPullSecrets.GetRegistryURL())
		return
	}
	r := correspondingRegistryOfImage.(*registry)
	r.addOrUpdateImage(ctx, obj)
}

func (rs *Registries) registryIsNotInStoreYet(registryURL string) bool {
	if _, found := rs.store.Load(registryURL); !found {
		return true
	}
	return false
}

func (rs *Registries) addNewRegistry(registryURL string) {

	rs.store.Store(registryURL, nil)
}

func (rs *Registries) UpdateImage(ctx context.Context, obj event.ObservedKubernetesAPIObjectEvent) {
	registryURL := obj.ImageWithPullSecrets.GetRegistryURL()
	correspondingRegistryOfImage, found := rs.store.Load(registryURL)
	if !found {
		log.Errorf("tried to update image for object with uuid \"%s\" but corresponding registry \"%s\" does not exists", obj.GetUID(), obj.ImageWithPullSecrets.GetRegistryURL())
		return
	}
	r := correspondingRegistryOfImage.(*registry)
	r.addOrUpdateImage(ctx, obj)
}

func (rs *Registries) DeleteImage(obj event.ObservedKubernetesAPIObjectEvent) {
	registryURL := obj.ImageWithPullSecrets.GetRegistryURL()
	correspondingRegistryOfImage, found := rs.store.Load(registryURL)
	if !found {
		log.Errorf("tried to delete image for object with uuid \"%s\" but corresponding registry \"%s\" does not exists", obj.GetUID(), obj.ImageWithPullSecrets.GetRegistryURL())
		return
	}
	r := correspondingRegistryOfImage.(*registry)
	r.deleteImage(obj)
}
