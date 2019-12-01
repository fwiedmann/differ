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

package appv1scraper

import (
	"github.com/fwiedmann/differ/pkg/kubernetes-scraper/util"
	"github.com/fwiedmann/differ/pkg/opts"
	"github.com/fwiedmann/differ/pkg/store"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Deployment type struct
type Deployment struct {
}

// todo: implement watch method with chanels
// GetWorkloadResources scrapes all appsV1 deployments
func (d Deployment) GetWorkloadResources(c *kubernetes.Clientset, namespace string, resourceStore *store.Instance) error {
	deployments, err := c.AppsV1().Deployments(namespace).List(v1.ListOptions{})
	if err != nil {
		return err
	}

	for _, deployment := range deployments.Items {
		if _, ok := deployment.Annotations[opts.DifferAnnotation]; ok {
			authSecrets, err := util.GetRegistryAuth(deployment.Spec.Template.Spec.ImagePullSecrets, c, namespace)
			if err != nil {
				return err
			}
			resourceStore.AddResource("appsV1", "Deployment", deployment.Namespace, deployment.Name, deployment.Spec.Template.Spec.Containers, authSecrets)
		}
	}
	return nil
}
