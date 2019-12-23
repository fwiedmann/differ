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

package util

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/watch"

	"github.com/fwiedmann/differ/pkg/types"

	"github.com/fwiedmann/differ/pkg/store"

	log "github.com/sirupsen/logrus"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
)

type ImageWithAssociatedPullSecrets struct {
	ImageName        string
	ImageTag         string
	ImagePullSecrets []store.ImagePullSecret
}

func getRegistryAuth(ImagePullSecretNames []string, obseverCondfig *types.KubernetesObserverConfig) (map[string][]store.ImagePullSecret, error) {
	marshaledAuths := make(map[string][]store.ImagePullSecret)

	for _, secretName := range ImagePullSecretNames {
		secret, err := obseverCondfig.KubernetesAPIClient.CoreV1().Secrets(obseverCondfig.NamespaceToScrape).Get(secretName, metaV1.GetOptions{
			TypeMeta: metaV1.TypeMeta{
				Kind: "kubernetes.io/dockerconfigjson",
			},
		})
		if err != nil {
			log.Error(err)
			continue
		}

		// using map interface for unmarshal because of arbitrary keys in the dockerconfigjson
		jsonContent := make(map[string]interface{})
		err = json.Unmarshal(secret.Data[".dockerconfigjson"], &jsonContent)
		if err != nil {
			log.Error(err)
			continue
		}

		for registry, auths := range jsonContent["auths"].(map[string]interface{}) {
			auth := store.ImagePullSecret{}
			for jsonAuthKey, jsonAuthValue := range auths.(map[string]interface{}) {

				switch jsonAuthKey {
				case "username":
					auth.Username = jsonAuthValue.(string)
				case "password":
					auth.Password = jsonAuthValue.(string)
				}
			}
			marshaledAuths[registry] = append(marshaledAuths[registry], auth)
		}
	}
	return marshaledAuths, nil
}

func GetImagesAndImagePullSecrets(pod v1.PodTemplateSpec, observerConfig *types.KubernetesObserverConfig) ([]ImageWithAssociatedPullSecrets, error) {

	images := getImagesFromPodSpec(pod.Spec)
	imagePullSecretNames := getImagePullSecretNamesFromPodSpec(pod.Spec)

	imagePullSecrets, err := getRegistryAuth(imagePullSecretNames, observerConfig)
	if err != nil {
		return []ImageWithAssociatedPullSecrets{}, nil
	}
	return generateImageAssociatedAndPullSecretsCombinations(images, imagePullSecrets), nil

}

func generateImageAssociatedAndPullSecretsCombinations(images []string, imagePullSecrets map[string][]store.ImagePullSecret) []ImageWithAssociatedPullSecrets {
	imageSecretsCombinations := make([]ImageWithAssociatedPullSecrets, 0)

	for _, image := range images {
		imageAssociatedPullSecrets := make([]store.ImagePullSecret, 0)
		for registryName, secret := range imagePullSecrets {
			if imageBelongsToRegistry(image, registryName) {
				imageAssociatedPullSecrets = append(imageAssociatedPullSecrets, secret...)
			}
		}
		imageName, imageTag := separateImageAndTag(image)
		imageSecretsCombinations = append(imageSecretsCombinations, ImageWithAssociatedPullSecrets{
			ImageName:        imageName,
			ImageTag:         imageTag,
			ImagePullSecrets: imageAssociatedPullSecrets,
		})
	}
	return imageSecretsCombinations
}

func imageBelongsToRegistry(image string, registry string) bool {
	if strings.Contains(image, registry) || !strings.Contains(image, ".") {
		return true
	}
	return false
}
func getImagesFromPodSpec(pod v1.PodSpec) []string {
	images := []string{}
	for _, container := range pod.Containers {
		images = append(images, container.Image)
	}
	return images
}

func getImagePullSecretNamesFromPodSpec(pod v1.PodSpec) []string {
	imagePullSecretNames := []string{}
	for _, secret := range pod.ImagePullSecrets {
		imagePullSecretNames = append(imagePullSecretNames, secret.Name)
	}
	return imagePullSecretNames
}

func separateImageAndTag(image string) (string, string) {
	separatedImage := strings.Split(image, ":")
	if len(separatedImage) == 2 {
		return separatedImage[0], separatedImage[1]
	}
	return image, "latest"
}

func ApiObjectEventToString(event watch.Event) string {
	return fmt.Sprintf("%s", event.Type)
}
