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
	"fmt"

	"k8s.io/client-go/informers"
)

const (
	AppV1Deployment  = Kind("APP_V1_DEPLOYMENT")
	AppV1DaemonSet   = Kind("APP_V1_DAEMON_SET")
	AppV1StatefulSet = Kind("APP_V1_STATEFUL_SET")
)

type Kind string

// NewObserver init an observer for the given type with corresponding kubernetesObjectHandler.
// If the observer kind could not be found will return error
func NewObserver(observerKind Kind, observerConfig Config) (*Observer, error) {
	switch observerKind {
	case AppV1Deployment:
		return newAppsV1DeploymentObserver(observerConfig), nil
	case AppV1DaemonSet:
		return newAppsV1DaemonSetObserver(observerConfig), nil
	case AppV1StatefulSet:
		return newAppsV1StatefulSetObserver(observerConfig), nil
	}
	return nil, fmt.Errorf("observer kind %s not found", observerKind)
}

func initNewKubernetesFactory(observerConfig Config) informers.SharedInformerFactory {
	return informers.NewSharedInformerFactoryWithOptions(observerConfig.kubernetesAPIClient, 0, informers.WithNamespace(observerConfig.namespaceToScrape))
}
