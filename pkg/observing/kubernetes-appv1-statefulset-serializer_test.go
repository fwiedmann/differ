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

func createStatefulSet(name, kind, uid string, pod v1.PodTemplateSpec) *appsV1.StatefulSet {

	return &appsV1.StatefulSet{
		TypeMeta: metaV1.TypeMeta{
			Kind: kind,
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			UID:       types.UID(uid),
			Namespace: "default",
		},
		Spec: appsV1.StatefulSetSpec{
			Template: pod,
		},
	}
}

func TestNewKubernetesAPPV1StatefulSetSerializer(t *testing.T) {
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
				kubernetesAPIObj: createStatefulSet("test1", "StatefulSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: KubernetesAPPV1StatefulSetSerializer{
				convertedStatefulSet: createStatefulSet("test1", "StatefulSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			wantErr: false,
		},
		{
			name: "InvalidObject",
			args: args{
				kubernetesAPIObj: appsV1.Deployment{},
			},
			want: KubernetesAPPV1StatefulSetSerializer{
				convertedStatefulSet: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewKubernetesAPPV1StatefulSetSerializer(tt.args.kubernetesAPIObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKubernetesAPPV1StatefulSetSerializer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKubernetesAPPV1StatefulSetSerializer() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1StatefulSetSerializer_GetObjectKind(t *testing.T) {
	type fields struct {
		convertedStatefulSet *appsV1.StatefulSet
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedStatefulSet: createStatefulSet("test1", "StatefulSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "StatefulSet",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statefulSetObjectSerializer := KubernetesAPPV1StatefulSetSerializer{
				convertedStatefulSet: tt.fields.convertedStatefulSet,
			}
			if got := statefulSetObjectSerializer.GetObjectKind(); got != tt.want {
				t.Errorf("GetObjectKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1StatefulSetSerializer_GetName(t *testing.T) {
	type fields struct {
		convertedStatefulSet *appsV1.StatefulSet
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedStatefulSet: createStatefulSet("test1", "StatefulSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "test1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statefulSetObjectSerializer := KubernetesAPPV1StatefulSetSerializer{
				convertedStatefulSet: tt.fields.convertedStatefulSet,
			}
			if got := statefulSetObjectSerializer.GetName(); got != tt.want {
				t.Errorf("GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1StatefulSetSerializer_GetAPIVersion(t *testing.T) {
	type fields struct {
		convertedStatefulSet *appsV1.StatefulSet
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedStatefulSet: createStatefulSet("test1", "StatefulSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "appV1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statefulSetObjectSerializer := KubernetesAPPV1StatefulSetSerializer{
				convertedStatefulSet: tt.fields.convertedStatefulSet,
			}
			if got := statefulSetObjectSerializer.GetAPIVersion(); got != tt.want {
				t.Errorf("GetAPIVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1StatefulSetSerializer_GetPodSpec(t *testing.T) {
	type fields struct {
		convertedStatefulSet *appsV1.StatefulSet
	}
	tests := []struct {
		name   string
		fields fields
		want   v1.PodSpec
	}{
		{
			name: "Valid",
			fields: fields{
				convertedStatefulSet: createStatefulSet("test1", "StatefulSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: createPodSpecTemplate("test1", "differ").Spec,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statefulSetObjectSerializer := KubernetesAPPV1StatefulSetSerializer{
				convertedStatefulSet: tt.fields.convertedStatefulSet,
			}
			if got := statefulSetObjectSerializer.GetPodSpec(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPodSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1StatefulSetSerializer_GetUID(t *testing.T) {
	type fields struct {
		convertedStatefulSet *appsV1.StatefulSet
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedStatefulSet: createStatefulSet("test1", "StatefulSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "187",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statefulSetObjectSerializer := KubernetesAPPV1StatefulSetSerializer{
				convertedStatefulSet: tt.fields.convertedStatefulSet,
			}
			if got := statefulSetObjectSerializer.GetUID(); got != tt.want {
				t.Errorf("GetUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesAPPV1StatefulSetSerializer_GetNamespace(t *testing.T) {
	type fields struct {
		convertedStatefulSet *appsV1.StatefulSet
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid",
			fields: fields{
				convertedStatefulSet: createStatefulSet("test1", "StatefulSet", "187", createPodSpecTemplate("test1", "differ")),
			},
			want: "default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statefulSetObjectSerializer := KubernetesAPPV1StatefulSetSerializer{
				convertedStatefulSet: tt.fields.convertedStatefulSet,
			}
			if got := statefulSetObjectSerializer.GetNamespace(); got != tt.want {
				t.Errorf("GetNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
