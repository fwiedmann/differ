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

package observing

import (
	"errors"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
)

func NewKubernetesAPPV1DaemonSetSerializer(kubernetesAPIObj interface{}) (KubernetesObjectSerializer, error) {
	convertedDaemonSet, err := convertToDaemonSet(kubernetesAPIObj)
	if err != nil {
		return KubernetesAPPV1DaemonSetSerializer{}, err
	}
	return KubernetesAPPV1DaemonSetSerializer{convertedDaemonSet: convertedDaemonSet}, nil
}

func convertToDaemonSet(kubernetesAPIObj interface{}) (*appsV1.DaemonSet, error) {
	convertedDaemonSet, ok := kubernetesAPIObj.(*appsV1.DaemonSet)
	if !ok {
		return nil, errors.New("observing/KubernetesAPPV1DaemonSetSerializer error: could not parse apps/appsV1 DaemonSet object")
	}
	return convertedDaemonSet, nil
}

// KubernetesAPPV1DaemonSetSerializer for kubernetes appV1/DaemonSet
type KubernetesAPPV1DaemonSetSerializer struct {
	convertedDaemonSet *appsV1.DaemonSet
}

func (daemonSetObjectSerializer KubernetesAPPV1DaemonSetSerializer) GetObjectKind() string {
	return "DaemonSet"
}

func (daemonSetObjectSerializer KubernetesAPPV1DaemonSetSerializer) GetName() string {
	return daemonSetObjectSerializer.convertedDaemonSet.GetName()
}

func (daemonSetObjectSerializer KubernetesAPPV1DaemonSetSerializer) GetAPIVersion() string {
	return "appV1"
}

// GetPodSpec from appV1/DaemonSet Object
func (daemonSetObjectSerializer KubernetesAPPV1DaemonSetSerializer) GetPodSpec() coreV1.PodSpec {
	return daemonSetObjectSerializer.convertedDaemonSet.Spec.Template.Spec
}

func (daemonSetObjectSerializer KubernetesAPPV1DaemonSetSerializer) GetUID() string {
	return string(daemonSetObjectSerializer.convertedDaemonSet.GetUID())
}
