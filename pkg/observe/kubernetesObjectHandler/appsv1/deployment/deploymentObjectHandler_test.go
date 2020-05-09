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

package deployment

import (
	"reflect"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsV1 "k8s.io/api/apps/v1"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var validDeployment = &appsV1.Deployment{

	ObjectMeta: v1.ObjectMeta{
		Name: "test-deployment-set",
		UID:  "7fd28ba2-0a08-4dc8-ad4c-df336906fc94",
	},
	Spec: appsV1.DeploymentSpec{
		Template: coreV1.PodTemplateSpec{
			Spec: coreV1.PodSpec{
				Containers: []coreV1.Container{
					{
						Name:  "test-name-0",
						Image: "test-image-0",
					},
					{
						Name:  "test-name-1",
						Image: "test-image-1",
					},
				},
			},
		},
	},
}

func TestHandler_GetNameOfObservedObject(t *testing.T) {
	type fields struct {
		convertedDeployment *appsV1.Deployment
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "WithValidName",
			fields: fields{convertedDeployment: validDeployment},
			want:   validDeployment.GetName(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploymentObjectHandler := Handler{
				convertedDeployment: tt.fields.convertedDeployment,
			}
			if got := deploymentObjectHandler.GetNameOfObservedObject(); got != tt.want {
				t.Errorf("GetNameOfObservedObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_GetPodSpec(t *testing.T) {
	type fields struct {
		convertedDeployment *appsV1.Deployment
	}
	tests := []struct {
		name   string
		fields fields
		want   coreV1.PodSpec
	}{
		{
			name:   "WithValidPodSpec",
			fields: fields{convertedDeployment: validDeployment},
			want:   validDeployment.Spec.Template.Spec,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploymentObjectHandler := Handler{
				convertedDeployment: tt.fields.convertedDeployment,
			}
			if got := deploymentObjectHandler.GetPodSpec(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPodSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_GetUID(t *testing.T) {
	type fields struct {
		convertedDeployment *appsV1.Deployment
	}
	tests := []struct {
		name   string
		fields fields
		want   types.UID
	}{
		{
			name:   "WithValidUID",
			fields: fields{convertedDeployment: validDeployment},
			want:   validDeployment.GetUID(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploymentObjectHandler := Handler{
				convertedDeployment: tt.fields.convertedDeployment,
			}
			if got := deploymentObjectHandler.GetUID(); got != tt.want {
				t.Errorf("GetUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewHandler(t *testing.T) {
	type args struct {
		kubernetesAPIObj interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    Handler
		wantErr bool
	}{
		{
			name:    "WithValidKubernetesObject",
			args:    args{kubernetesAPIObj: validDeployment},
			want:    Handler{convertedDeployment: validDeployment},
			wantErr: false,
		},
		{
			name:    "WithInvalidKubernetesObject",
			args:    args{kubernetesAPIObj: &appsV1.DaemonSet{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHandler(tt.args.kubernetesAPIObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToDeployment(t *testing.T) {
	type args struct {
		kubernetesAPIObj interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *appsV1.Deployment
		wantErr bool
	}{
		{
			name:    "ValidKubernetesObject",
			args:    args{kubernetesAPIObj: validDeployment},
			want:    validDeployment,
			wantErr: false,
		},
		{
			name:    "InvalidKubernetesObject",
			args:    args{kubernetesAPIObj: &appsV1.DaemonSet{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToDeployment(tt.args.kubernetesAPIObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToDeployment() got = %v, want %v", got, tt.want)
			}
		})
	}
}
