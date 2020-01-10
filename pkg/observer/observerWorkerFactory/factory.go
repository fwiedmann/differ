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

package observerWorkerFactory

import (
	"github.com/fwiedmann/differ/pkg/observer"
	"github.com/fwiedmann/differ/pkg/observer/observerWorkerFactory/appv1/daemonSet"
	"github.com/fwiedmann/differ/pkg/observer/observerWorkerFactory/appv1/deployment"
	"github.com/fwiedmann/differ/pkg/observer/observerWorkerFactory/appv1/statefulSet"
	"k8s.io/client-go/informers"
)

const (
	AppV1Deployment  = ObserverWorkerType("APP_V1_DEPLOYMENT")
	AppV1DaemonSet   = ObserverWorkerType("APP_V1_DAEMON_SET")
	AppV1StatefulSet = ObserverWorkerType("APP_V1_STATEFUL_SET")
)

type ObserverWorkerType string

// NewObserverWorker init an ObserverWorker for the given type. Returns nil if the given type does not exists
func NewObserverWorker(observerType ObserverWorkerType, config observer.Config) observer.Worker {
	kubernetesFactory := initNewKubernetesFactory(config)
	switch observerType {
	case AppV1Deployment:
		return deployment.InitObserverWorker(kubernetesFactory)
	case AppV1DaemonSet:
		return daemonSet.InitObserverWorker(kubernetesFactory)
	case AppV1StatefulSet:
		return statefulSet.InitObserverWorker(kubernetesFactory)
	}
	return nil
}

func initNewKubernetesFactory(observerConfig observer.Config) informers.SharedInformerFactory {
	return informers.NewSharedInformerFactoryWithOptions(observerConfig.KubernetesAPIClient, 0, informers.WithNamespace(observerConfig.NamespaceToScrape))
}
