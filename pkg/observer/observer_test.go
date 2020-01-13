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
	"reflect"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/fwiedmann/differ/pkg/event"
	"k8s.io/client-go/kubernetes"
)

func newFakeKubernetesClient() kubernetes.Interface {
	return fake.NewSimpleClientset()
}
func newFakeObserverConfig() Config {
	testNamespace := "default"
	fakeKubernetesClient := newFakeKubernetesClient()

	return Config{
		namespaceToScrape:                    testNamespace,
		kubernetesAPIClient:                  fakeKubernetesClient,
		eventGenerator:                       event.NewGenerator(fakeKubernetesClient, testNamespace),
		KubernetesEventCommunicationChannels: event.NewCommunicationChannels(1),
	}
}

func TestNewObserverConfig(t *testing.T) {
	fakeKubernetesClient := newFakeKubernetesClient()
	testNamespace := "default"
	testCommunicationChannels := event.NewCommunicationChannels(1)
	testEventGenerator := event.NewGenerator(fakeKubernetesClient, testNamespace)

	type args struct {
		namespaceToScrape                    string
		kubernetesAPIClient                  kubernetes.Interface
		kubernetesEventCommunicationChannels event.KubernetesEventCommunicationChannels
		eventGenerator                       *event.Generator
	}
	tests := []struct {
		name string
		args args
		want Config
	}{
		{
			name: "WithValidConfigArguments",
			args: args{
				namespaceToScrape:                    testNamespace,
				kubernetesAPIClient:                  fakeKubernetesClient,
				kubernetesEventCommunicationChannels: testCommunicationChannels,
				eventGenerator:                       testEventGenerator,
			},
			want: Config{
				kubernetesAPIClient:                  fakeKubernetesClient,
				eventGenerator:                       testEventGenerator,
				namespaceToScrape:                    testNamespace,
				KubernetesEventCommunicationChannels: testCommunicationChannels,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewObserverConfig(tt.args.namespaceToScrape, tt.args.kubernetesAPIClient, tt.args.kubernetesEventCommunicationChannels, tt.args.eventGenerator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewObserverConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObserver_StartObserving(t *testing.T) {

	validTestObserverConfig := newFakeObserverConfig()
	validObserver, err := NewObserver(AppV1Deployment, validTestObserverConfig)
	if err != nil {
		t.Fatalf("NewObserver() failed for StartObserving()")
	}

	tests := []struct {
		name               string
		testObserver       *Observer
		testDeploymentName string
	}{
		{
			name:               "WithValidObserver",
			testObserver:       validObserver,
			testDeploymentName: "test-deployment",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go tt.testObserver.StartObserving()
			time.Sleep(time.Second * 2)
			tt.testObserver.StopObserving()
		})
	}
}

func TestObserver_StopObserving(t *testing.T) {
	validTestObserverConfig := newFakeObserverConfig()
	validObserver, err := NewObserver(AppV1Deployment, validTestObserverConfig)
	if err != nil {
		t.Fatalf("NewObserver() failed for StartObserving()")
	}

	tests := []struct {
		name         string
		testObserver *Observer
	}{
		{
			name:         "WithValidObserver",
			testObserver: validObserver,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go tt.testObserver.StartObserving()
			time.Sleep(time.Second * 2)
			tt.testObserver.StopObserving()

			if _, channelIsOpen := <-tt.testObserver.stopChannel; channelIsOpen {
				t.Fatalf("StopObserving() did not close the stop chan")
			}
		})
	}
}
