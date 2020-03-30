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

package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/fwiedmann/differ/pkg/image"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Generator struct {
	KubernetesAPIClient kubernetes.Interface
	WorkingNamespace    string
}

// NewGenerator instance which will transform a Kubernetes api object to events for the communication channels
func NewGenerator(kubernetesAPIClient kubernetes.Interface, workingNamespace string) *Generator {
	return &Generator{
		KubernetesAPIClient: kubernetesAPIClient,
		WorkingNamespace:    workingNamespace}
}

// GenerateEventsFromPodSpec will extract all containers from the podSpec, requests all given image pull secrets from the kubernetes API
// and generate events for each container
func (eventGenerator *Generator) GenerateEventsFromPodSpec(podSpec v1.PodSpec, kubernetesMetaInformation KubernetesAPIObjectMetaInformation) ([]ObservedKubernetesAPIObjectEvent, error) {
	extractedImagesFromPodSpec := eventGenerator.extractImagesFromPodSpec(podSpec)
	extractedPullSecretsFromPodSpec, err := eventGenerator.extractPullSecretsFromPodSpec(podSpec)

	if err != nil {
		return []ObservedKubernetesAPIObjectEvent{}, err
	}
	updatedImagesWithPullSecrets := appendPullSecretsWhichBelongsToImage(extractedImagesFromPodSpec, extractedPullSecretsFromPodSpec)

	return createEventForEachImage(updatedImagesWithPullSecrets, kubernetesMetaInformation), nil
}

func (eventGenerator *Generator) extractImagesFromPodSpec(pod v1.PodSpec) []image.WithAssociatedPullSecrets {
	var images []image.WithAssociatedPullSecrets
	for _, container := range pod.Containers {
		images = append(images, image.NewWithAssociatedPullSecrets(container.Image, container.Name))
	}
	return images
}

func (eventGenerator *Generator) extractPullSecretsFromPodSpec(pod v1.PodSpec) (map[string][]image.PullSecret, error) {
	imagePullSecretNames := extractNamesOfImagePullSecretFromPodSpec(pod)
	return eventGenerator.getAllImagePullSecretsByRegistry(imagePullSecretNames)
}

func extractNamesOfImagePullSecretFromPodSpec(pod v1.PodSpec) []string {
	var imagePullSecretNames []string
	for _, secret := range pod.ImagePullSecrets {
		imagePullSecretNames = append(imagePullSecretNames, secret.Name)
	}
	return imagePullSecretNames
}

func (eventGenerator *Generator) getAllImagePullSecretsByRegistry(ImagePullSecretNames []string) (map[string][]image.PullSecret, error) {
	allImagePullSecretsFromPodPerRegistry := make(map[string][]image.PullSecret)

	for _, secretName := range ImagePullSecretNames {
		secretFromAPI, err := eventGenerator.getImagePullSecretFromAPI(secretName)
		if err != nil {
			return make(map[string][]image.PullSecret), err
		}

		unmarshalledSecret, err := unmarshalSecret(secretFromAPI)
		if err != nil {
			return make(map[string][]image.PullSecret), err
		}
		appendImagePullSecretsToRegistry(unmarshalledSecret, allImagePullSecretsFromPodPerRegistry)
	}
	return allImagePullSecretsFromPodPerRegistry, nil
}

func (eventGenerator *Generator) getImagePullSecretFromAPI(secretName string) (*v1.Secret, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	secret, err := eventGenerator.KubernetesAPIClient.CoreV1().Secrets(eventGenerator.WorkingNamespace).Get(ctx, secretName, metaV1.GetOptions{
		TypeMeta: metaV1.TypeMeta{
			Kind: "kubernetes.io/dockerconfigjson",
		},
	})

	cancel()
	return secret, err
}

func unmarshalSecret(secret *v1.Secret) (map[string]interface{}, error) {
	unmarshalledSecret := make(map[string]interface{})
	err := json.Unmarshal(secret.Data[".dockerconfigjson"], &unmarshalledSecret)
	return unmarshalledSecret, err
}

func appendImagePullSecretsToRegistry(unmarshalledSecret map[string]interface{}, allSecretsFromPod map[string][]image.PullSecret) {
	for registry, auth := range unmarshalledSecret["auths"].(map[string]interface{}) {
		pullSecret := getImagePullSecretFromRegistryInterface(auth)
		allSecretsFromPod[registry] = append(allSecretsFromPod[registry], pullSecret)
	}
}

func getImagePullSecretFromRegistryInterface(auth interface{}) image.PullSecret {
	pullSecret := image.PullSecret{}
	for jsonAuthKey, jsonAuthValue := range auth.(map[string]interface{}) {
		switch jsonAuthKey {
		case "username":
			pullSecret.Username = jsonAuthValue.(string)
		case "password":
			pullSecret.Password = jsonAuthValue.(string)
		}
	}
	return pullSecret
}

func appendPullSecretsWhichBelongsToImage(images []image.WithAssociatedPullSecrets, allPullSecrets map[string][]image.PullSecret) []image.WithAssociatedPullSecrets {
	updatedImages := make([]image.WithAssociatedPullSecrets, 0)
	for _, extractedImage := range images {
		extractedImage.AppendImagePullSecretsWhichBelongsToImage(allPullSecrets)
		updatedImages = append(updatedImages, extractedImage)
	}
	return updatedImages
}

func createEventForEachImage(images []image.WithAssociatedPullSecrets, kubernetesMetaInformation KubernetesAPIObjectMetaInformation) (generatedEvents []ObservedKubernetesAPIObjectEvent) {
	for _, imageForEvent := range images {
		generatedEvents = append(generatedEvents, ObservedKubernetesAPIObjectEvent{
			MetaInformation:      kubernetesMetaInformation,
			ImageWithPullSecrets: imageForEvent,
		})
	}
	return
}
