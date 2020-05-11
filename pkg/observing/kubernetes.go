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
	"context"
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/tools/cache"

	"github.com/fwiedmann/differ/pkg/differentiating"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubernetesObjectSerializer interface {
	GetPodSpec() v1.PodSpec
	GetObjectKind() string
	GetName() string
	GetUID() string
	GetAPIVersion() string
}

type KubernetesObserverService struct {
	ds         differentiating.Service
	client     kubernetes.Interface
	namespace  string
	serializer func(obj interface{}) (KubernetesObjectSerializer, error)
}

func StartKubernetesObserverService(ctx context.Context, c kubernetes.Interface, informer cache.SharedInformer, ns string, objSerializer func(obj interface{}) (KubernetesObjectSerializer, error), service differentiating.Service) *KubernetesObserverService {
	kos := &KubernetesObserverService{
		ds:         service,
		client:     c,
		namespace:  ns,
		serializer: objSerializer,
	}

	stop := make(chan struct{})

	informer.AddEventHandler(kos)
	go informer.Run(stop)

	serviceCtx, cancel := context.WithCancel(ctx)
	go func(ctx context.Context) {
		<-ctx.Done()
		cancel()
		stop <- struct{}{}
	}(serviceCtx)

	return kos
}

func (k *KubernetesObserverService) OnAdd(obj interface{}) {
	k.handleInformerEvent("adding", obj, k.ds.AddImage)
}

func (k *KubernetesObserverService) OnUpdate(_, newObj interface{}) {
	k.handleInformerEvent("update", newObj, k.ds.UpdateImage)
}

func (k *KubernetesObserverService) OnDelete(obj interface{}) {
	k.handleInformerEvent("deletion", obj, k.ds.DeleteImage)
}

func (k *KubernetesObserverService) handleInformerEvent(operationKind string, kubernetesObj interface{}, differntiateServiceOperation func(ctx context.Context, i differentiating.Image) error) {
	o, err := k.serializer(kubernetesObj)
	if err != nil {
		log.Errorf("observing/kubernetes error: %s", err)
		return
	}

	images, err := k.getImagesFromPodSpec(o.GetPodSpec(), kubernetesAPIObjectMetaInformation{
		UID:          o.GetUID(),
		APIVersion:   o.GetAPIVersion(),
		ResourceType: o.GetObjectKind(),
		Namespace:    k.namespace,
		WorkloadName: o.GetName(),
	})
	if err != nil {
		log.Errorf("observing/kubernetes error: %s", err)
		return
	}

	for _, kubernetesImage := range images {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		err := make(chan error)

		go func(i imageWithKubernetesMetadata) {
			ps := make([]*differentiating.PullSecret, 0)
			for _, p := range i.Image.pullSecrets {
				ps = append(ps, &differentiating.PullSecret{
					Username: p.username,
					Password: p.password,
				})
				log.Infof("%v", p)
			}

			err <- differntiateServiceOperation(ctx, differentiating.Image{
				ID:       i.GetUID(),
				Registry: i.Image.GetRegistryURL(),
				Name:     i.Image.GetNameWithoutRegistry(),
				Tag:      i.Image.GetTag(),
				Auth:     ps,
			})
		}(kubernetesImage)

		select {
		case err := <-err:
			if err != nil {
				log.Errorf("observing/kubernetes error: %s", err)
			}
			cancel()
			return
		case <-ctx.Done():
			cancel()
			log.Errorf("observing/kubernetes error: could not perform %s action on differentiate service, timout exceeded", operationKind)
			return
		}
	}
}

func (k *KubernetesObserverService) getImagesFromPodSpec(podSpec v1.PodSpec, kubernetesMetaInformation kubernetesAPIObjectMetaInformation) ([]imageWithKubernetesMetadata, error) {
	extractedImagesFromPodSpec := k.extractImagesFromPodSpec(podSpec)
	extractedPullSecretsFromPodSpec, err := k.extractPullSecretsFromPodSpec(podSpec)

	if err != nil {
		return []imageWithKubernetesMetadata{}, err
	}
	updatedImagesWithPullSecrets := appendPullSecretsWhichBelongsToImage(extractedImagesFromPodSpec, extractedPullSecretsFromPodSpec)

	return createEventForEachImage(updatedImagesWithPullSecrets, kubernetesMetaInformation), nil
}

func (k *KubernetesObserverService) extractImagesFromPodSpec(pod v1.PodSpec) []image {
	var images []image
	for _, container := range pod.Containers {
		image, err := NewImage(container.Image, container.Name)
		if err != nil {
			log.Error(err)
			continue
		}
		images = append(images, image)
	}
	return images
}

func (k *KubernetesObserverService) extractPullSecretsFromPodSpec(pod v1.PodSpec) (map[string][]*pullSecret, error) {
	imagePullSecretNames := extractNamesOfImagePullSecretFromPodSpec(pod)
	return k.getAllImagePullSecretsByRegistry(imagePullSecretNames)
}

func extractNamesOfImagePullSecretFromPodSpec(pod v1.PodSpec) []string {
	var imagePullSecretNames []string
	for _, secret := range pod.ImagePullSecrets {
		imagePullSecretNames = append(imagePullSecretNames, secret.Name)
	}
	return imagePullSecretNames
}

func (k *KubernetesObserverService) getAllImagePullSecretsByRegistry(ImagePullSecretNames []string) (map[string][]*pullSecret, error) {
	allImagePullSecretsFromPodPerRegistry := make(map[string][]*pullSecret)

	for _, secretName := range ImagePullSecretNames {
		secretFromAPI, err := k.getImagePullSecretFromAPI(secretName)
		if err != nil {
			return make(map[string][]*pullSecret), err
		}

		unmarshalledSecret, err := unmarshalSecret(secretFromAPI)
		if err != nil {
			return make(map[string][]*pullSecret), err
		}
		appendImagePullSecretsToRegistry(unmarshalledSecret, allImagePullSecretsFromPodPerRegistry)
	}
	return allImagePullSecretsFromPodPerRegistry, nil
}

func (k *KubernetesObserverService) getImagePullSecretFromAPI(secretName string) (*v1.Secret, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	secret, err := k.client.CoreV1().Secrets(k.namespace).Get(ctx, secretName, metaV1.GetOptions{
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

func appendImagePullSecretsToRegistry(unmarshalledSecret map[string]interface{}, allSecretsFromPod map[string][]*pullSecret) {
	for registry, auth := range unmarshalledSecret["auths"].(map[string]interface{}) {
		pullSecret := getImagePullSecretFromRegistryInterface(auth)
		allSecretsFromPod[registry] = append(allSecretsFromPod[registry], pullSecret)
	}
}

func getImagePullSecretFromRegistryInterface(auth interface{}) *pullSecret {
	var username, password string
	for jsonAuthKey, jsonAuthValue := range auth.(map[string]interface{}) {
		switch jsonAuthKey {
		case "username":
			username = jsonAuthValue.(string)
		case "password":
			password = jsonAuthValue.(string)
		}
	}
	return newPullSecret(username, password)
}

func appendPullSecretsWhichBelongsToImage(images []image, allPullSecrets map[string][]*pullSecret) []image {
	updatedImages := make([]image, 0)
	for _, extractedImage := range images {
		extractedImage.AppendImagePullSecretsWhichBelongsToImage(allPullSecrets)
		updatedImages = append(updatedImages, extractedImage)
	}
	return updatedImages
}

func createEventForEachImage(images []image, kubernetesMetaInformation kubernetesAPIObjectMetaInformation) (generatedEvents []imageWithKubernetesMetadata) {
	for _, imageForEvent := range images {
		generatedEvents = append(generatedEvents, imageWithKubernetesMetadata{
			MetaInformation: kubernetesMetaInformation,
			Image:           imageForEvent,
		})
	}
	return
}
