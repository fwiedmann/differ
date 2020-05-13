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
	"reflect"
	"testing"

	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func createDeployment(name, kind, uid string, pod v1.PodTemplateSpec) *appsV1.Deployment {

	return &appsV1.Deployment{
		TypeMeta: metaV1.TypeMeta{
			Kind: kind,
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name: name,
			UID:  types.UID(uid),
		},
		Spec: appsV1.DeploymentSpec{
			Template: pod,
		},
	}
}

func TestNewKubernetesAPPV1DeploymentSerializer(t *testing.T) {
	type args struct {
		kubernetesAPIObj interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    KubernetesObjectSerializer
		wantErr bool
	}{
		{
			name: "ValidObject",
			args: args{
				kubernetesAPIObj: createDeployment("test1", "Deployment", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: KubernetesAPPV1DeploymentSerializer{
				convertedDeployment: createDeployment("test1", "Deployment", "187", createPodSpecTemplate("test1", "differ")),
			},
			wantErr: false,
		},
		{
			name: "InvalidObject",
			args: args{
				kubernetesAPIObj: appsV1.StatefulSet{},
			},
			want: KubernetesAPPV1DeploymentSerializer{
				convertedDeployment: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewKubernetesAPPV1DeploymentSerializer(tt.args.kubernetesAPIObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKubernetesAPPV1DeploymentSerializer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKubernetesAPPV1DeploymentSerializer() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1DeploymentSerializer_GetObjectKind(t *testing.T) {
	type fields struct {
		convertedDeployment *appsV1.Deployment
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedDeployment: createDeployment("test1", "Deployment", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "Deployment",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploymentObjectSerializer := KubernetesAPPV1DeploymentSerializer{
				convertedDeployment: tt.fields.convertedDeployment,
			}
			if got := deploymentObjectSerializer.GetObjectKind(); got != tt.want {
				t.Errorf("GetObjectKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1DeploymentSerializer_GetName(t *testing.T) {
	type fields struct {
		convertedDeployment *appsV1.Deployment
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedDeployment: createDeployment("test1", "Deployment", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "test1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploymentObjectSerializer := KubernetesAPPV1DeploymentSerializer{
				convertedDeployment: tt.fields.convertedDeployment,
			}
			if got := deploymentObjectSerializer.GetName(); got != tt.want {
				t.Errorf("GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1DeploymentSerializer_GetAPIVersion(t *testing.T) {
	type fields struct {
		convertedDeployment *appsV1.Deployment
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedDeployment: createDeployment("test1", "Deployment", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "appV1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploymentObjectSerializer := KubernetesAPPV1DeploymentSerializer{
				convertedDeployment: tt.fields.convertedDeployment,
			}
			if got := deploymentObjectSerializer.GetAPIVersion(); got != tt.want {
				t.Errorf("GetAPIVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1DeploymentSerializer_GetPodSpec(t *testing.T) {
	type fields struct {
		convertedDeployment *appsV1.Deployment
	}
	tests := []struct {
		name   string
		fields fields
		want   v1.PodSpec
	}{
		{
			name: "Valid",
			fields: fields{
				convertedDeployment: createDeployment("test1", "Deployment", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: createPodSpecTemplate("test1", "differ").Spec,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploymentObjectSerializer := KubernetesAPPV1DeploymentSerializer{
				convertedDeployment: tt.fields.convertedDeployment,
			}
			if got := deploymentObjectSerializer.GetPodSpec(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPodSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1DeploymentSerializer_GetUID(t *testing.T) {
	type fields struct {
		convertedDeployment *appsV1.Deployment
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedDeployment: createDeployment("test1", "Deployment", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "187",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploymentObjectSerializer := KubernetesAPPV1DeploymentSerializer{
				convertedDeployment: tt.fields.convertedDeployment,
			}
			if got := deploymentObjectSerializer.GetUID(); got != tt.want {
				t.Errorf("GetUID() = %v, want %v", got, tt.want)
			}
		})
	}
}
