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

// NewHandler try's to convert the kubernetes API Object to an *appsV1.StatefulSet.
// If conversion is not successful will return error.
func NewKubernetesAPPV1StatefulSetSerializer(kubernetesAPIObj interface{}) (KubernetesObjectSerializer, error) {
	convertedStatefulSet, err := convertToStateFulSet(kubernetesAPIObj)
	if err != nil {
		return KubernetesAPPV1StatefulSetSerializer{}, err
	}
	return KubernetesAPPV1StatefulSetSerializer{convertedStatefulSet: convertedStatefulSet}, nil
}

func convertToStateFulSet(kubernetesAPIObj interface{}) (*appsV1.StatefulSet, error) {
	convertedStatefulSet, ok := kubernetesAPIObj.(*appsV1.StatefulSet)
	if !ok {
		return nil, errors.New("observing/KubernetesAPPV1StatefulSetSerializer error:could not parse apps/appsV1 StatefulSet object")
	}
	return convertedStatefulSet, nil
}

// KubernetesAPPV1StatefulSetSerializer for kubernetes appV1/StatefulSet
type KubernetesAPPV1StatefulSetSerializer struct {
	convertedStatefulSet *appsV1.StatefulSet
}

func (statefulSetObjectSerializer KubernetesAPPV1StatefulSetSerializer) GetObjectKind() string {
	return "StatefulSet"
}

func (statefulSetObjectSerializer KubernetesAPPV1StatefulSetSerializer) GetName() string {
	return statefulSetObjectSerializer.convertedStatefulSet.GetName()
}

func (statefulSetObjectSerializer KubernetesAPPV1StatefulSetSerializer) GetAPIVersion() string {
	return "appV1"
}

// GetPodSpec from appV1/StatefulSet Object
func (statefulSetObjectSerializer KubernetesAPPV1StatefulSetSerializer) GetPodSpec() coreV1.PodSpec {
	return statefulSetObjectSerializer.convertedStatefulSet.Spec.Template.Spec
}

// GetUID from appV1/StatefulSet Object
func (statefulSetObjectSerializer KubernetesAPPV1StatefulSetSerializer) GetUID() string {
	return string(statefulSetObjectSerializer.convertedStatefulSet.GetUID())
}
