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

package kubernetesscraper

import (
	"errors"

	"github.com/fwiedmann/differ/pkg/kubernetesscraper/appv1scraper"
	"github.com/fwiedmann/differ/pkg/types"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func init() {
	apiObservers = append(apiObservers, appv1scraper.NewDeploymentObserver())
}

var (
	apiObservers []KubernetesAPIObserver
)

type KubernetesAPIObserver interface {
	InitObserverWithConfig(observeConfig *types.KubernetesObserverConfig) error
	SendObservedKubernetesAPIResource()
	UpdateConfig(*types.KubernetesObserverConfig) error
}

type KubernetesAPISObserver struct {
	ObserverChannel chan types.ObservedImageEvent
}

func InitKubernetesAPIObserversWithKubernetesClient(namespaceToScrape string) (*KubernetesAPISObserver, error) {

	kubernetesAPIClient, err := initKubernetesClient()
	if err != nil {
		return nil, err
	}

	observerChannel := make(chan types.ObservedImageEvent, len(apiObservers))

	err = initAllObservers(&types.KubernetesObserverConfig{
		NamespaceToScrape:   namespaceToScrape,
		KubernetesAPIClient: kubernetesAPIClient,
		ObserverChannel:     observerChannel,
	})

	if err != nil {
		return nil, err
	}

	return &KubernetesAPISObserver{
		ObserverChannel: observerChannel,
	}, nil
}

func (o *KubernetesAPISObserver) ObserveAPIs() {
	for _, kubernetesAPIObserver := range apiObservers {
		go kubernetesAPIObserver.SendObservedKubernetesAPIResource()
	}
}

func initKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return &kubernetes.Clientset{}, err
	}

	if clientset, err := kubernetes.NewForConfig(config); err != nil {
		return &kubernetes.Clientset{}, err
	} else {
		return clientset, nil
	}
}
func initAllObservers(observeConfig *types.KubernetesObserverConfig) error {
	var initFailed bool
	for _, kubernetesAPIObserver := range apiObservers {

		if err := kubernetesAPIObserver.InitObserverWithConfig(observeConfig); err != nil {
			log.Error(err)
			initFailed = true
		}
	}
	if initFailed {
		return errors.New("could not initialize all kubernetes api servers")
	}
	return nil
}
