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
	"errors"

	coreV1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/api/apps/v1"
)

func newDeploymentObjectHandler(obj interface{}) (Handler, error) {
	daemonSet, err := convertToDeployment(obj)
	if err != nil {
		return Handler{}, err
	}
	return Handler{convertedDaemonSet: daemonSet}, nil
}
func convertToDeployment(obj interface{}) (*v1.Deployment, error) {
	deployment, ok := obj.(*v1.Deployment)
	if !ok {
		return nil, errors.New("could not parse Deployment object")
	}
	return deployment, nil
}

// Handler for kubernetes appV1/Deployment
type Handler struct {
	convertedDaemonSet *v1.Deployment
}

// GetPodSpec from appV1/Deployment Object
func (deploymentObjectHandler Handler) GetPodSpec() coreV1.PodSpec {
	return deploymentObjectHandler.convertedDaemonSet.Spec.Template.Spec
}

// GetNameOfObservedObject from appV1/Deployment Object
func (deploymentObjectHandler Handler) GetNameOfObservedObject() string {
	return deploymentObjectHandler.convertedDaemonSet.Name
}

// GetUID from appV1/Deployment Object
func (deploymentObjectHandler Handler) GetUID() types.UID {
	return deploymentObjectHandler.convertedDaemonSet.GetUID()
}
