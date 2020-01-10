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

package deployment

import (
	"github.com/fwiedmann/differ/pkg/observer"

	"k8s.io/client-go/informers"

	"k8s.io/client-go/tools/cache"
)

// Deployment worker for the kubernetes API appsV1/Deployment type
type Deployment struct {
	kubernetesSharedInformer cache.SharedIndexInformer
	objectKind               string
	apiVersion               string
}

// InitObserverWorker for the appV1/Deployment worker
func InitObserverWorker(kubernetesFactory informers.SharedInformerFactory) *Deployment {
	return &Deployment{
		kubernetesSharedInformer: kubernetesFactory.Apps().V1().Deployments().Informer(),
		objectKind:               "Deployment",
		apiVersion:               "appsV1",
	}
}

// NewHandlerForObject will init a worker handler for each new object
func (deploymentObserverWorker *Deployment) NewHandlerForObject(obj interface{}) (observer.ObjectHandler, error) {
	return newDeploymentObjectHandler(obj)
}

// GetAPIVersion of the kubernetes appV1/Deployment
func (deploymentObserverWorker *Deployment) GetAPIVersion() string {
	return deploymentObserverWorker.apiVersion
}

// GetObservedAPIObjectKind of the kubernetes API Object
func (deploymentObserverWorker *Deployment) GetObservedAPIObjectKind() string {
	return deploymentObserverWorker.objectKind
}

// GetSharedIndexInformer of the kubernetes appV1/Deployment sharedIndexInformer
func (deploymentObserverWorker *Deployment) GetSharedIndexInformer() cache.SharedIndexInformer {
	return deploymentObserverWorker.kubernetesSharedInformer
}
