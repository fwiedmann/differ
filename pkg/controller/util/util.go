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
	"reflect"
	"regexp"
	"sort"

	"github.com/fwiedmann/differ/pkg/store"

	"github.com/fwiedmann/differ/pkg/registry"
	log "github.com/sirupsen/logrus"
)

// todo: add further image tag patterns
var imageTagPatterns = []string{
	"^[0-9].[0-9].[0-9]$",
}

func IsValidTag(tag string) (bool, string) {
	for _, pattern := range imageTagPatterns {
		valid, err := regexp.Match(pattern, []byte(tag))
		if err != nil {
			log.Error(err)
		}
		if valid {
			return valid, pattern
		}
	}
	return false, ""
}

func SortTagsByPattern(tags []string, pattern string) []string {
	var validTags []string
	for _, tag := range tags {
		if valid, _ := regexp.Match(pattern, []byte(tag)); valid {
			validTags = append(validTags, tag)
		}
	}
	sort.Strings(validTags)
	return validTags
}

func IsRegistryError(err error) error {
	if err, ok := err.(registry.Error); ok {
		log.Warn(err)
		return nil
	}
	return err
}

func GatherAuths(resourceInfos []store.KubernetesAPIResource) []store.ImagePullSecret {
	auths := make([]store.ImagePullSecret, 0)

	for _, resourceInfo := range resourceInfos {
		for _, secret := range resourceInfo.Secrets {
			var exists bool
			for _, storedSecret := range auths {
				if reflect.DeepEqual(secret, storedSecret) {
					exists = true
				}
			}
			if !exists {
				auths = append(auths, secret)
			}
		}
	}
	return auths
}
