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

	"github.com/fwiedmann/differ/pkg/store"

	log "github.com/sirupsen/logrus"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func GetRegistryAuth(secretRefs []core.LocalObjectReference, client *kubernetes.Clientset, namespace string) (map[string][]store.ImagePullSecret, error) {
	marshaledAuths := make(map[string][]store.ImagePullSecret)

	for _, secretRef := range secretRefs {
		secret, err := client.CoreV1().Secrets(namespace).Get(secretRef.Name, v1.GetOptions{
			TypeMeta: v1.TypeMeta{
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
