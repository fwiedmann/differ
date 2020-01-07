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
	"github.com/fwiedmann/differ/pkg/event"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

type EventGenerator interface {
	GenerateEventsFromPodSpec(podSpec v1.PodSpec, kubernetesMetaInformation event.KubernetesAPIObjectMetaInformation) ([]event.ObservedKubernetesAPIObjectEvent, error)
}

type Config struct {
	NamespaceToScrape   string
	KubernetesAPIClient kubernetes.Interface
	event.KubernetesEventCommunicationChannels
	EventGenerator *event.Generator
}

func InitNewKubernetesFactory(observerConfig Config) informers.SharedInformerFactory {
	return informers.NewSharedInformerFactoryWithOptions(observerConfig.KubernetesAPIClient, 0, informers.WithNamespace(observerConfig.NamespaceToScrape))
}
