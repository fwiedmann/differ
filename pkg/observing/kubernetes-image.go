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
	"fmt"
)

// imageWithKubernetesMetadata contains unique meta information from scraped resource types
type imageWithKubernetesMetadata struct {
	MetaInformation kubernetesAPIObjectMetaInformation
	Image           image
}

// GetUID generates a unique ID based of the kubernetes metadata
func (o imageWithKubernetesMetadata) GetUID() string {
	return fmt.Sprintf("%s_%s_%s_%s_%s_%s", o.MetaInformation.Namespace, o.MetaInformation.APIVersion, o.MetaInformation.UID, o.MetaInformation.WorkloadName, o.MetaInformation.ResourceType, o.Image.GetContainerName())
}

// String implements the stringer interface
func (o imageWithKubernetesMetadata) String() string {
	return fmt.Sprintf("MetaInformation: %s, image: %s", o.MetaInformation, o.Image)
}

// kubernetesAPIObjectMetaInformation from the kubernetes API object
type kubernetesAPIObjectMetaInformation struct {
	UID          string
	APIVersion   string
	ResourceType string
	Namespace    string
	WorkloadName string
}

// String implements the stringer interface
func (k kubernetesAPIObjectMetaInformation) String() string {
	return fmt.Sprintf("UID: %s, APIVersion: %s, ResourceType: %s, Namespace: %s, WorkloadName: %s", k.UID, k.APIVersion, k.ResourceType, k.Namespace, k.WorkloadName)
}
