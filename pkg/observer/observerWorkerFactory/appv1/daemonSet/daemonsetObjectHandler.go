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

package daemonSet

import (
	"errors"

	coreV1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/api/apps/v1"
)

func newDaemonSetHandler(obj interface{}) (Handler, error) {
	daemonSet, err := convertToDaemonSet(obj)
	if err != nil {
		return Handler{}, err
	}
	return Handler{convertedDaemonSet: daemonSet}, nil
}
func convertToDaemonSet(obj interface{}) (*v1.DaemonSet, error) {
	daemonSet, ok := obj.(*v1.DaemonSet)
	if !ok {
		return nil, errors.New("could not parse DaemonSet object")
	}
	return daemonSet, nil
}

// Handler for kubernetes appV1/DaemonSet
type Handler struct {
	convertedDaemonSet *v1.DaemonSet
}

// GetPodSpec from appV1/DaemonSet Object
func (daemonSetObjectHandler Handler) GetPodSpec() coreV1.PodSpec {
	return daemonSetObjectHandler.convertedDaemonSet.Spec.Template.Spec
}

// GetNameOfObservedObject from appV1/DaemonSet Object
func (daemonSetObjectHandler Handler) GetNameOfObservedObject() string {
	return daemonSetObjectHandler.convertedDaemonSet.Name
}

// GetUID from appV1/DaemonSet Object
func (daemonSetObjectHandler Handler) GetUID() types.UID {
	return daemonSetObjectHandler.convertedDaemonSet.GetUID()
}
