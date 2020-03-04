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

package image

import (
	"fmt"
	"strconv"
	"strings"
)

type WithAssociatedPullSecrets struct {
	containerName string
	imageName     string
	imageTag      string
	pullSecrets   []PullSecret
}

type PullSecret struct {
	Username string
	Password string
}

func NewWithAssociatedPullSecrets(rawImage, containerName string) WithAssociatedPullSecrets {
	name, tag := separateImageAndTag(rawImage)
	return WithAssociatedPullSecrets{
		containerName: containerName,
		imageName:     name,
		imageTag:      tag,
	}
}

func (i *WithAssociatedPullSecrets) GetContainerName() string {
	return i.containerName
}

func (i *WithAssociatedPullSecrets) GetName() string {
	return i.imageName
}

func (i *WithAssociatedPullSecrets) GetTag() string {
	return i.imageTag
}

func (i *WithAssociatedPullSecrets) GetPullSecrets() []PullSecret {
	return i.pullSecrets
}

func (i *WithAssociatedPullSecrets) GetRegistryURL() string {
	separatedURLAndImage := strings.Split(i.containerName, "/")
	return separatedURLAndImage[0]
}

func (i *WithAssociatedPullSecrets) AppendImagePullSecretsWhichBelongsToImage(pullSecrets map[string][]PullSecret) {
	for registryName, secrets := range pullSecrets {
		if imageBelongsToRegistry(i.GetName(), registryName) {
			i.appendPullSecrets(secrets)
		}
	}
}

func (i *WithAssociatedPullSecrets) appendPullSecrets(matchedPullSecrets []PullSecret) {
	i.pullSecrets = append(i.pullSecrets, matchedPullSecrets...)

}
func imageBelongsToRegistry(image string, registry string) bool {
	if strings.Contains(image, registry) {
		return true
	}
	return false
}

func separateImageAndTag(rawImage string) (imageName string, imageTag string) {
	separatedImage := splitImage(rawImage)

	if isDockerHubImage(separatedImage[0]) {
		separatedImage[0] = "docker.io/" + separatedImage[0]
	}
	switch {
	case hasPortAndTag(separatedImage):
		return fmt.Sprintf("%s:%s", separatedImage[0], separatedImage[1]), separatedImage[2]
	case hasOnlyTag(separatedImage):
		return separatedImage[0], separatedImage[1]
	case hasOnlyPort(separatedImage):
		return fmt.Sprintf("%s:%s", separatedImage[0], separatedImage[1]), "latest"
	}
	return separatedImage[0], "latest"
}

func splitImage(image string) []string {
	return strings.Split(image, ":")
}

func isDockerHubImage(imageName string) bool {
	return !strings.Contains(imageName, ".")
}

func hasOnlyTag(separatedImage []string) bool {
	if len(separatedImage) != 2 {
		return false
	}
	if _, err := strconv.Atoi(separatedImage[1]); err == nil {
		return false
	}
	return true
}
func hasPortAndTag(separatedImage []string) bool {
	return len(separatedImage) == 3
}
func hasOnlyPort(separatedImage []string) bool {
	if len(separatedImage) != 2 {
		return false
	}
	if _, err := strconv.Atoi(separatedImage[1]); err != nil {
		return false
	}
	return true
}
