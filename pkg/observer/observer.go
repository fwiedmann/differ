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

package observer

import (
	"errors"

	"github.com/fwiedmann/differ/pkg/observer/appv1scraper"
	"github.com/fwiedmann/differ/pkg/types"
	log "github.com/sirupsen/logrus"
)

var (
	apiObservers []KubernetesAPIObserver
)

func init() {
	apiObservers = append(apiObservers, appv1scraper.NewDeploymentObserver(), appv1scraper.NewDaemonSetObserver(), appv1scraper.NewStatefulSetObserver())
}

// KubernetesAPISObserver represents the basic behaviour of observers
type KubernetesAPIObserver interface {
	InitObserverWithConfig(observeConfig *types.KubernetesObserverConfig) error
	SendObservedKubernetesAPIResource()
	UpdateConfig(*types.KubernetesObserverConfig) error
}

// KubernetesAPISObserver initialize all api scraper and provides an event channel
type KubernetesAPISObserver struct {
	ObserverChannel chan types.ObservedImageEvent
}

// InitKubernetesAPIObservers with types.KubernetesObserverConfig
func InitKubernetesAPIObservers(kubernetesAPIClient types.KubernetesAPIClient) (*KubernetesAPISObserver, error) {
	observerConfig := newObserverConfig(kubernetesAPIClient)

	if err := initKubernetesAPIObservers(observerConfig); err != nil {
		return nil, err
	}
	return &KubernetesAPISObserver{
		ObserverChannel: observerConfig.ObserverChannel,
	}, nil
}

func newObserverConfig(kubernetesAPIClient types.KubernetesAPIClient) *types.KubernetesObserverConfig {
	return &types.KubernetesObserverConfig{
		ObserverChannel:   make(chan types.ObservedImageEvent, len(apiObservers)),
		NamespaceToScrape: kubernetesAPIClient.GetNameSpace(),
		KubernetesAPI:     kubernetesAPIClient,
	}

}

func initKubernetesAPIObservers(observerConfig *types.KubernetesObserverConfig) error {
	var initFailed bool
	for _, kubernetesAPIObserver := range apiObservers {
		if err := kubernetesAPIObserver.InitObserverWithConfig(observerConfig); err != nil {
			log.Error(err)
			initFailed = true
		}
	}
	if initFailed {
		return errors.New("could not initialize all kubernetes api servers")
	}
	return nil
}

// ObserveAPIs start all observers
func (kubernetesObserver *KubernetesAPISObserver) ObserveAPIs() {
	for _, kubernetesAPIObserver := range apiObservers {
		go kubernetesAPIObserver.SendObservedKubernetesAPIResource()
	}
}
