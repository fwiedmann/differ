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

	"k8s.io/apimachinery/pkg/types"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsV1 "k8s.io/api/apps/v1"

	v1 "k8s.io/api/core/v1"
)

func createDaemonSet(name, kind, uid string, pod v1.PodTemplateSpec) *appsV1.DaemonSet {
	return &appsV1.DaemonSet{
		TypeMeta: metaV1.TypeMeta{
			Kind: kind,
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			UID:       types.UID(uid),
			Namespace: "default",
		},
		Spec: appsV1.DaemonSetSpec{
			Template: pod,
		},
	}
}

func createPodSpecTemplate(containerName, image string) v1.PodTemplateSpec {
	return v1.PodTemplateSpec{
		ObjectMeta: metaV1.ObjectMeta{},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: containerName, Image: image},
			},
		},
	}
}

func TestNewKubernetesAPPV1DaemonSetSerializer(t *testing.T) {
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
				kubernetesAPIObj: createDaemonSet("test1", "DaemonSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: KubernetesAPPV1DaemonSetSerializer{
				convertedDaemonSet: createDaemonSet("test1", "DaemonSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			wantErr: false,
		},
		{
			name: "InvalidObject",
			args: args{
				kubernetesAPIObj: appsV1.StatefulSet{},
			},
			want: KubernetesAPPV1DaemonSetSerializer{
				convertedDaemonSet: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewKubernetesAPPV1DaemonSetSerializer(tt.args.kubernetesAPIObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKubernetesAPPV1DaemonSetSerializer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKubernetesAPPV1DaemonSetSerializer() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1DaemonSetSerializer_GetObjectKind(t *testing.T) {
	type fields struct {
		convertedDaemonSet *appsV1.DaemonSet
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedDaemonSet: createDaemonSet("test1", "DaemonSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "DaemonSet",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonSetObjectSerializer := KubernetesAPPV1DaemonSetSerializer{
				convertedDaemonSet: tt.fields.convertedDaemonSet,
			}
			if got := daemonSetObjectSerializer.GetObjectKind(); got != tt.want {
				t.Errorf("GetObjectKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1DaemonSetSerializer_GetName(t *testing.T) {
	type fields struct {
		convertedDaemonSet *appsV1.DaemonSet
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedDaemonSet: createDaemonSet("test1", "DaemonSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "test1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonSetObjectSerializer := KubernetesAPPV1DaemonSetSerializer{
				convertedDaemonSet: tt.fields.convertedDaemonSet,
			}
			if got := daemonSetObjectSerializer.GetName(); got != tt.want {
				t.Errorf("GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1DaemonSetSerializer_GetAPIVersion(t *testing.T) {
	type fields struct {
		convertedDaemonSet *appsV1.DaemonSet
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedDaemonSet: createDaemonSet("test1", "DaemonSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "appV1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonSetObjectSerializer := KubernetesAPPV1DaemonSetSerializer{
				convertedDaemonSet: tt.fields.convertedDaemonSet,
			}
			if got := daemonSetObjectSerializer.GetAPIVersion(); got != tt.want {
				t.Errorf("GetAPIVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1DaemonSetSerializer_GetPodSpec(t *testing.T) {
	type fields struct {
		convertedDaemonSet *appsV1.DaemonSet
	}
	tests := []struct {
		name   string
		fields fields
		want   v1.PodSpec
	}{
		{
			name: "Valid",
			fields: fields{
				convertedDaemonSet: createDaemonSet("test1", "DaemonSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: createPodSpecTemplate("test1", "differ").Spec,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonSetObjectSerializer := KubernetesAPPV1DaemonSetSerializer{
				convertedDaemonSet: tt.fields.convertedDaemonSet,
			}
			if got := daemonSetObjectSerializer.GetPodSpec(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPodSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1DaemonSetSerializer_GetUID(t *testing.T) {
	type fields struct {
		convertedDaemonSet *appsV1.DaemonSet
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedDaemonSet: createDaemonSet("test1", "DaemonSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "187",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonSetObjectSerializer := KubernetesAPPV1DaemonSetSerializer{
				convertedDaemonSet: tt.fields.convertedDaemonSet,
			}
			if got := daemonSetObjectSerializer.GetUID(); got != tt.want {
				t.Errorf("GetUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1DaemonSetSerializer_GetNamespace(t *testing.T) {
	type fields struct {
		convertedDaemonSet *appsV1.DaemonSet
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedDaemonSet: createDaemonSet("test1", "DaemonSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonSetObjectSerializer := KubernetesAPPV1DaemonSetSerializer{
				convertedDaemonSet: tt.fields.convertedDaemonSet,
			}
			if got := daemonSetObjectSerializer.GetNamespace(); got != tt.want {
				t.Errorf("GetNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
