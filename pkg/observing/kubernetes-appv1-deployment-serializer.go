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

func NewKubernetesAPPV1DeploymentSerializer(kubernetesAPIObj interface{}) (KubernetesObjectSerializer, error) {
	convertedDaemonSet, err := convertToDeployment(kubernetesAPIObj)
	if err != nil {
		return KubernetesAPPV1DeploymentSerializer{}, err
	}
	return KubernetesAPPV1DeploymentSerializer{convertedDeployment: convertedDaemonSet}, nil
}

func convertToDeployment(kubernetesAPIObj interface{}) (*appsV1.Deployment, error) {
	convertedDeployment, ok := kubernetesAPIObj.(*appsV1.Deployment)
	if !ok {
		return nil, errors.New("observing/KubernetesAPPV1DeploymentSerializer error:could not parse  apps/appsV1 Deployment object")
	}
	return convertedDeployment, nil
}

// KubernetesAPPV1DeploymentSerializer for kubernetes appV1/Deployment
type KubernetesAPPV1DeploymentSerializer struct {
	convertedDeployment *appsV1.Deployment
}

func (deploymentObjectSerializer KubernetesAPPV1DeploymentSerializer) GetObjectKind() string {
	return "Deployment"
}

func (deploymentObjectSerializer KubernetesAPPV1DeploymentSerializer) GetName() string {
	return deploymentObjectSerializer.convertedDeployment.GetName()
}

func (deploymentObjectSerializer KubernetesAPPV1DeploymentSerializer) GetAPIVersion() string {
	return "appV1"
}

// GetPodSpec from appV1/Deployment Object
func (deploymentObjectSerializer KubernetesAPPV1DeploymentSerializer) GetPodSpec() coreV1.PodSpec {
	return deploymentObjectSerializer.convertedDeployment.Spec.Template.Spec
}

// GetUID from appV1/Deployment Object
func (deploymentObjectSerializer KubernetesAPPV1DeploymentSerializer) GetUID() string {
	return string(deploymentObjectSerializer.convertedDeployment.GetUID())
}
