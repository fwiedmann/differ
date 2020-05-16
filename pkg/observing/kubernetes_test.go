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
	"fmt"
	"testing"
	"time"

	"k8s.io/client-go/informers"

	"k8s.io/client-go/kubernetes/fake"

	coreV1 "k8s.io/api/core/v1"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fwiedmann/differ/pkg/differentiating"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	testImage           = "wiedmannfelix/differ"
	usernameAndPassword = "admin"
	pullSecretName      = "test-pull-validSecret"
	testNamespace       = "test-name-space"
	validSecret         = &coreV1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      pullSecretName,
			Namespace: testNamespace,
		},
		Type: coreV1.DockerConfigJsonKey,
		Data: map[string][]byte{
			coreV1.DockerConfigJsonKey: []byte(fmt.Sprintf("{\"auths\":{\"docker.io\":{\"username\":\"%s\",\"password\":\"%s\",\"email\":\"admin@example.com\",\"auth\":\"YWRtaW46YWRtaW4K\"}}}", usernameAndPassword, usernameAndPassword)),
		},
	}
	invalidSecret = &coreV1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      pullSecretName,
			Namespace: testNamespace,
		},
		Type: coreV1.DockerConfigJsonKey,
		Data: map[string][]byte{},
	}
)

func createUpdateDeleteObjects(t *testing.T, ctx context.Context, client kubernetes.Interface) {
	workloadContext, cancel := context.WithCancel(ctx)
	var count int
	for {
		select {
		case <-workloadContext.Done():
			cancel()
			return
		case <-time.After(time.Minute * 2):
			cancel()
			return
		default:
			count++
			obj := &v1.Deployment{
				ObjectMeta: metaV1.ObjectMeta{
					Name: fmt.Sprintf("deployment-%d", count),
				},
				Spec: v1.DeploymentSpec{
					Template: coreV1.PodTemplateSpec{
						Spec: coreV1.PodSpec{
							ImagePullSecrets: []coreV1.LocalObjectReference{{Name: pullSecretName}},
							Containers: []coreV1.Container{
								{
									Name:  fmt.Sprintf("container-1-%d", count),
									Image: testImage,
								},
								{
									Name:  fmt.Sprintf("container-2-%d", count),
									Image: testImage,
								},
								{
									Name:  fmt.Sprintf("container-3-%d", count),
									Image: testImage,
								},
								{
									Name:  fmt.Sprintf("container-4-%d", count),
									Image: testImage,
								},
							},
						},
					},
				},
			}

			_, err := client.AppsV1().Deployments(testNamespace).Create(workloadContext, obj, metaV1.CreateOptions{})
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(time.Microsecond * 100)
			_, err = client.AppsV1().Deployments(testNamespace).Update(workloadContext, obj, metaV1.UpdateOptions{})
			if err != nil {
				t.Fatal(err)
			}

			time.Sleep(time.Microsecond * 100)
			err = client.AppsV1().Deployments(testNamespace).Delete(workloadContext, obj.Name, metaV1.DeleteOptions{})
			if err != nil {
				t.Fatal(err)
			}
		}

	}
}

func TestStartKubernetesObserverService(t *testing.T) {
	type args struct {
		c                kubernetes.Interface
		ns               string
		objSerializer    func(obj interface{}) (KubernetesObjectSerializer, error)
		wantServiceError bool
		createWorkload   func(t *testing.T, ctx context.Context, client kubernetes.Interface)
	}
	tests := []struct {
		name string
		args args
		want *KubernetesObserverService
	}{
		{
			name: "Valid",
			args: args{
				c:                fake.NewSimpleClientset(validSecret),
				ns:               testNamespace,
				objSerializer:    NewKubernetesAPPV1DeploymentSerializer,
				wantServiceError: false,
				createWorkload:   createUpdateDeleteObjects,
			},
			want: &KubernetesObserverService{
				ds:         differentiating.MockService{},
				namespace:  testNamespace,
				serializer: NewKubernetesAPPV1DeploymentSerializer,
			},
		},
		{
			name: "PullSecretNotFound",
			args: args{
				c:                fake.NewSimpleClientset(),
				ns:               testNamespace,
				objSerializer:    NewKubernetesAPPV1DeploymentSerializer,
				wantServiceError: false,
				createWorkload:   createUpdateDeleteObjects,
			},
			want: &KubernetesObserverService{
				ds:         differentiating.MockService{},
				namespace:  testNamespace,
				serializer: NewKubernetesAPPV1DeploymentSerializer,
			},
		},
		{
			name: "InvalidPullSecret",
			args: args{
				c:                fake.NewSimpleClientset(invalidSecret),
				ns:               testNamespace,
				objSerializer:    NewKubernetesAPPV1DeploymentSerializer,
				wantServiceError: false,
				createWorkload:   createUpdateDeleteObjects,
			},
			want: &KubernetesObserverService{
				ds:         differentiating.MockService{},
				namespace:  testNamespace,
				serializer: NewKubernetesAPPV1DeploymentSerializer,
			},
		},
		{
			name: "DifferentiateServiceError",
			args: args{
				c:                fake.NewSimpleClientset(validSecret),
				ns:               testNamespace,
				objSerializer:    NewKubernetesAPPV1DeploymentSerializer,
				wantServiceError: true,
				createWorkload:   createUpdateDeleteObjects,
			},
			want: &KubernetesObserverService{

				namespace:  testNamespace,
				serializer: NewKubernetesAPPV1DeploymentSerializer,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			validateFun := func(i differentiating.Image) error {
				if i.Name != testImage {
					t.Errorf("StartKubernetesObserverService() = %v, want %v\", got", testImage, i.Name)
				}
				if i.Auth[0].Password != usernameAndPassword {
					t.Errorf("StartKubernetesObserverService() = %v, want %v\", got", usernameAndPassword, i.Auth[0].Password)
				}
				if i.Auth[0].Username != usernameAndPassword {
					t.Errorf("StartKubernetesObserverService() = %v, want %v\", got", usernameAndPassword, i.Auth[0].Username)
				}
				return nil
			}

			if tt.args.wantServiceError {
				validateFun = func(i differentiating.Image) error {
					return fmt.Errorf("differentiating/service mock error")
				}
			}

			differentiatingServiceMock := differentiating.MockService{
				Add:    validateFun,
				Delete: validateFun,
				Update: validateFun,
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			i := informers.NewSharedInformerFactoryWithOptions(tt.args.c, 0, informers.WithNamespace(testNamespace)).Apps().V1().Deployments().Informer()

			defer cancel()
			tt.want.client = tt.args.c
			StartKubernetesObserverService(ctx, tt.args.c, i, tt.args.ns, tt.args.objSerializer, differentiatingServiceMock)

			go tt.args.createWorkload(t, ctx, tt.args.c)
			<-ctx.Done()
		})
	}
}
