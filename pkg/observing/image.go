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
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	dockerHubURL = "registry-1.docker.io"
)

var (
	regexpDockerHubLibraryLatest  = regexp.MustCompile("^[\\-a-zA-Z]+$")
	regexpDockerHubLibraryWithTag = regexp.MustCompile("^[\\-a-zA-Z]+:[a-zA-z0-9\\.\\-\\_\\/]+$")
	regexpDockerHubLatest         = regexp.MustCompile("^[0-9\\-a-zA-Z\\/]+$")
	regexpDockerHubWithTag        = regexp.MustCompile("^[0-9\\-a-zA-Z\\/]+:[a-zA-z0-9\\.\\-\\_\\/]+$")
	regexpRegistryLatest          = regexp.MustCompile("^[a-zA-Z0-9\\.\\-]+.[a-z]+\\/[a-z\\/]+$")
	regexpRegistryWithTag         = regexp.MustCompile("^[a-zA-Z0-9\\.\\-]+.[a-z]+\\/[0-9\\-a-zA-Z\\/]+:[a-zA-z0-9\\.\\-\\_\\/]+$")
	regexpRegistryWithPortLatest  = regexp.MustCompile("^^[a-zA-Z0-9\\.\\-]+.[a-z]+:[0-9]+\\/[0-9\\-a-zA-Z\\/]+$")
	regexpRegistryWithPortWithTag = regexp.MustCompile("^^[a-zA-Z0-9\\.\\-]+.[a-z]+:[0-9]+\\/[0-9\\-a-zA-Z\\/]+:[a-zA-z0-9\\.\\-\\_\\/]+$")
)

type image struct {
	containerName       string
	name                string
	nameWithoutRegistry string
	registry            string
	tag                 string
	pullSecrets         []*pullSecret
}

func NewImage(rawImage, containerName string) (image, error) {

	if rawImage == "" {
		return image{}, fmt.Errorf("observing/imag error: container %s did not provide image name", containerName)
	}

	name, tag, err := separateImageAndTag(rawImage)
	if err != nil {
		logrus.Error(containerName)
		return image{}, err
	}

	return image{
		containerName:       containerName,
		name:                name,
		registry:            strings.Split(name, "/")[0],
		nameWithoutRegistry: getImageNameWithoutRegistry(name),
		tag:                 tag,
	}, nil
}

func (i image) String() string {
	return fmt.Sprintf("containerName: %s, name: %s, tag: %s, pullSecrets: %v", i.containerName, i.name, i.tag, i.pullSecrets)
}

func (i *image) GetContainerName() string {
	return i.containerName
}

func (i *image) GetNameWithRegistry() string {
	return i.name
}

func (i *image) GetNameWithoutRegistry() string {
	return i.nameWithoutRegistry
}

func (i *image) GetTag() string {
	return i.tag
}

func (i *image) GetPullSecrets() []*pullSecret {
	return i.pullSecrets
}

func (i *image) GetRegistryURL() string {
	return i.registry
}

func (i *image) AppendImagePullSecretsWhichBelongsToImage(pullSecrets map[string][]*pullSecret) {
	for registryName, secrets := range pullSecrets {
		if imageBelongsToRegistry(i.GetNameWithRegistry(), registryName) {
			i.pullSecrets = append(i.pullSecrets, secrets...)
		}
	}
}

func imageBelongsToRegistry(image string, registry string) bool {
	return strings.Contains(image, registry)
}

func separateImageAndTag(rawImage string) (imageName string, imageTag string, err error) {
	switch {
	case regexpDockerHubLibraryLatest.MatchString(rawImage):
		return fmt.Sprintf("%s/%s/%s", dockerHubURL, "library", rawImage), "latest", nil
	case regexpDockerHubLibraryWithTag.MatchString(rawImage):
		split := strings.Split(rawImage, ":")
		return fmt.Sprintf("%s/%s/%s", dockerHubURL, "library", split[0]), split[1], nil
	case regexpDockerHubLatest.MatchString(rawImage):
		return fmt.Sprintf("%s/%s", dockerHubURL, rawImage), "latest", nil
	case regexpDockerHubWithTag.MatchString(rawImage):
		split := strings.Split(rawImage, ":")
		return fmt.Sprintf("%s/%s", dockerHubURL, split[0]), split[1], nil
	case regexpRegistryLatest.MatchString(rawImage) || regexpRegistryWithPortLatest.MatchString(rawImage):
		return rawImage, "latest", nil
	case regexpRegistryWithTag.MatchString(rawImage):
		split := strings.Split(rawImage, ":")
		return split[0], split[1], nil
	case regexpRegistryWithPortWithTag.MatchString(rawImage):
		split := strings.Split(rawImage, ":")
		return fmt.Sprintf("%s:%s", split[0], split[1]), split[2], nil
	}
	return "", "", fmt.Errorf("observing/image error: could not analyze image %s, no valid regexepression found", rawImage)

}

func getImageNameWithoutRegistry(image string) string {
	imageParts := strings.Split(image, "/")

	var name string
	for i, part := range imageParts[1:] {
		if i == 0 {
			name = part
			continue
		}
		name = fmt.Sprintf("%s/%s", name, part)
	}
	return name
}
