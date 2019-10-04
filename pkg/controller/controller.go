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

package controller

import (
	"github.com/fwiedmann/differ/pkg/controller/util"
	"github.com/fwiedmann/differ/pkg/opts"
	"github.com/fwiedmann/differ/pkg/store"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

//ResourceScraper save scraped data in store
type ResourceScraper interface {
	GetWorkloadResources(c *kubernetes.Clientset, namespace string, scrapedResources store.Cache) error
}

// Controller type struct
type Controller struct {
	config *opts.ControllerConfig
}

// New initialize the differ controller
func New(c *opts.ControllerConfig) *Controller {
	return &Controller{
		config: c,
	}
}

// Run starts differ controller loop
func (c *Controller) Run(resourceScrapers []ResourceScraper) error {
	for {
		cache := make(store.Cache)

		kubernetesClient, err := util.InitKubernetesClient()
		if err != nil {
			return err
		}

		for _, s := range resourceScrapers {
			if err := s.GetWorkloadResources(kubernetesClient, c.config.Namespace, cache); err != nil {
				return err
			}
		}
		log.Debugf("Scraped resources:\n%v", cache)

		c.config.ControllerSleep()
	}
}
