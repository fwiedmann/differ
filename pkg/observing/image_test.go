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
)

func Test_image_GetContainerName(t *testing.T) {
	type fields struct {
		containerName string
		name          string
		tag           string
		pullSecrets   []*pullSecret
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Positive",
			fields: fields{
				containerName: "name",
			},
			want: "name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := image{
				containerName: tt.fields.containerName,
				name:          tt.fields.name,
				tag:           tt.fields.tag,
				pullSecrets:   tt.fields.pullSecrets,
			}
			if got := i.GetContainerName(); got != tt.want {
				t.Errorf("GetContainerName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_image_GetNameWithRegistry(t *testing.T) {
	type fields struct {
		containerName string
		name          string
		tag           string
		pullSecrets   []*pullSecret
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Positive",
			fields: fields{
				containerName: "name",
				name:          "gitlab.com/differ",
				tag:           "187",
				pullSecrets:   nil,
			},
			want: "gitlab.com/differ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := image{
				containerName: tt.fields.containerName,
				name:          tt.fields.name,
				tag:           tt.fields.tag,
				pullSecrets:   tt.fields.pullSecrets,
			}
			if got := i.GetNameWithRegistry(); got != tt.want {
				t.Errorf("GetNameWithRegistry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_image_GetTag(t *testing.T) {
	type fields struct {
		containerName       string
		name                string
		nameWithoutRegistry string
		registry            string
		tag                 string
		pullSecrets         []*pullSecret
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Positive",
			fields: fields{
				tag:         "1",
				pullSecrets: nil,
			},
			want: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := image{
				containerName:       tt.fields.containerName,
				name:                tt.fields.name,
				nameWithoutRegistry: tt.fields.nameWithoutRegistry,
				registry:            tt.fields.registry,
				tag:                 tt.fields.tag,
				pullSecrets:         tt.fields.pullSecrets,
			}
			if got := i.GetTag(); got != tt.want {
				t.Errorf("GetTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_image_GetPullSecrets(t *testing.T) {
	type fields struct {
		containerName       string
		name                string
		nameWithoutRegistry string
		registry            string
		tag                 string
		pullSecrets         []*pullSecret
	}
	tests := []struct {
		name   string
		fields fields
		want   []*pullSecret
	}{
		{
			name: "PositiveWithPullSecret",
			fields: fields{
				containerName:       "",
				name:                "",
				nameWithoutRegistry: "",
				registry:            "",
				tag:                 "",
				pullSecrets: []*pullSecret{
					{
						username: "admin",
						password: "admin",
					},
					{
						username: "superadmin",
						password: "superadmin",
					},
				},
			},
			want: []*pullSecret{
				{
					username: "admin",
					password: "admin",
				},
				{
					username: "superadmin",
					password: "superadmin",
				},
			},
		},
		{
			name:   "PositiveWithEmptyPullSecrets",
			fields: fields{},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := image{
				containerName:       tt.fields.containerName,
				name:                tt.fields.name,
				nameWithoutRegistry: tt.fields.nameWithoutRegistry,
				registry:            tt.fields.registry,
				tag:                 tt.fields.tag,
				pullSecrets:         tt.fields.pullSecrets,
			}
			if got := i.GetPullSecrets(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPullSecrets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_image_GetRegistryURL(t *testing.T) {
	type fields struct {
		containerName       string
		name                string
		nameWithoutRegistry string
		registry            string
		tag                 string
		pullSecrets         []*pullSecret
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Positive",
			fields: fields{
				registry: "gitlab.com",
			},
			want: "gitlab.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := image{
				containerName:       tt.fields.containerName,
				name:                tt.fields.name,
				nameWithoutRegistry: tt.fields.nameWithoutRegistry,
				registry:            tt.fields.registry,
				tag:                 tt.fields.tag,
				pullSecrets:         tt.fields.pullSecrets,
			}
			if got := i.GetRegistryURL(); got != tt.want {
				t.Errorf("GetRegistryURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_image_AppendImagePullSecretsWhichBelongsToImage(t *testing.T) {
	type fields struct {
		containerName       string
		name                string
		nameWithoutRegistry string
		registry            string
		tag                 string
		pullSecrets         []*pullSecret
	}
	type args struct {
		pullSecrets map[string][]*pullSecret
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		wantPullSecretsSize int
	}{
		{
			name: "Positive",
			fields: fields{
				containerName:       "",
				name:                "gitlab.com/wiedmannfelix/differ",
				nameWithoutRegistry: "wiedmannfelix/differ",
				registry:            "gitlab.com",
				tag:                 "",
				pullSecrets:         nil,
			},
			args: args{
				pullSecrets: map[string][]*pullSecret{
					"gitlab.com": {newPullSecret("admin", "admin"), newPullSecret("superadmin", "superadmin")},
					"github.com": {newPullSecret("admin", "admin")},
				},
			},
			wantPullSecretsSize: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := image{
				containerName:       tt.fields.containerName,
				name:                tt.fields.name,
				nameWithoutRegistry: tt.fields.nameWithoutRegistry,
				registry:            tt.fields.registry,
				tag:                 tt.fields.tag,
				pullSecrets:         tt.fields.pullSecrets,
			}
			i.AppendImagePullSecretsWhichBelongsToImage(tt.args.pullSecrets)

			if len(i.pullSecrets) != tt.wantPullSecretsSize {
				t.Errorf("AppendImagePullSecretsWhichBelongsToImage() error: pul")
			}
		})
	}
}

func TestNewImage1(t *testing.T) {
	type args struct {
		rawImage      string
		containerName string
	}
	tests := []struct {
		name    string
		args    args
		want    image
		wantErr bool
	}{
		{
			name: "DockerHubLibraryLatest",
			args: args{
				rawImage:      "differ",
				containerName: "container",
			},
			want: image{
				containerName:       "container",
				name:                dockerHubURL + "/library/differ",
				nameWithoutRegistry: "library/differ",
				registry:            dockerHubURL,
				tag:                 "latest",
				pullSecrets:         nil,
			},
			wantErr: false,
		},
		{
			name: "DockerHubLibraryWithTag",
			args: args{
				rawImage:      "differ:1.0.0",
				containerName: "container",
			},
			want: image{
				containerName:       "container",
				name:                dockerHubURL + "/library/differ",
				nameWithoutRegistry: "library/differ",
				registry:            dockerHubURL,
				tag:                 "1.0.0",
				pullSecrets:         nil,
			},
			wantErr: false,
		},
		{
			name: "DockerHubLatest",
			args: args{
				rawImage:      "wiedmann/differ",
				containerName: "container",
			},
			want: image{
				containerName:       "container",
				name:                dockerHubURL + "/wiedmann/differ",
				nameWithoutRegistry: "wiedmann/differ",
				registry:            dockerHubURL,
				tag:                 "latest",
				pullSecrets:         nil,
			},
			wantErr: false,
		},
		{
			name: "DockerHubWithTag",
			args: args{
				rawImage:      "wiedmann/differ:1.0.0",
				containerName: "container",
			},
			want: image{
				containerName:       "container",
				name:                dockerHubURL + "/wiedmann/differ",
				nameWithoutRegistry: "wiedmann/differ",
				registry:            dockerHubURL,
				tag:                 "1.0.0",
				pullSecrets:         nil,
			},
			wantErr: false,
		},
		{
			name: "RegistryLatest",
			args: args{
				rawImage:      "gitlab.com/wiedmann/differ",
				containerName: "container",
			},
			want: image{
				containerName:       "container",
				name:                "gitlab.com/wiedmann/differ",
				nameWithoutRegistry: "wiedmann/differ",
				registry:            "gitlab.com",
				tag:                 "latest",
				pullSecrets:         nil,
			},
		},
		{
			name: "RegistryWithTag",
			args: args{
				rawImage:      "gitlab.com/wiedmann/differ:1.0.0",
				containerName: "container",
			},
			want: image{
				containerName:       "container",
				name:                "gitlab.com/wiedmann/differ",
				nameWithoutRegistry: "wiedmann/differ",
				registry:            "gitlab.com",
				tag:                 "1.0.0",
				pullSecrets:         nil,
			},
		},
		{
			name: "RegistryWithPortLatest",
			args: args{
				rawImage:      "gitlab.com:8443/wiedmann/differ",
				containerName: "container",
			},
			want: image{
				containerName:       "container",
				name:                "gitlab.com:8443/wiedmann/differ",
				nameWithoutRegistry: "wiedmann/differ",
				registry:            "gitlab.com:8443",
				tag:                 "latest",
				pullSecrets:         nil,
			},
		},
		{
			name: "RegistryWithPortWithTag",
			args: args{
				rawImage:      "gitlab.com:8443/wiedmann/differ:1.0.0",
				containerName: "container",
			},
			want: image{
				containerName:       "container",
				name:                "gitlab.com:8443/wiedmann/differ",
				nameWithoutRegistry: "wiedmann/differ",
				registry:            "gitlab.com:8443",
				tag:                 "1.0.0",
				pullSecrets:         nil,
			},
		},
		{
			name: "InvalidName",
			args: args{
				rawImage:      "invalidURL.COM:%(",
				containerName: "container",
			},
			want: image{
				containerName:       "container",
				name:                "gitlab.com:8443/wiedmann/differ",
				nameWithoutRegistry: "wiedmann/differ",
				registry:            "gitlab.com:8443",
				tag:                 "1.0.0",
				pullSecrets:         nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewImage(tt.args.rawImage, tt.args.containerName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) && err == nil {
				t.Errorf("NewImage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_image_GetNameWithoutRegistry(t *testing.T) {
	type fields struct {
		containerName       string
		name                string
		nameWithoutRegistry string
		registry            string
		tag                 string
		pullSecrets         []*pullSecret
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Positive",
			fields: fields{
				nameWithoutRegistry: "differ/wiedmann",
			},
			want: "differ/wiedmann",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &image{
				containerName:       tt.fields.containerName,
				name:                tt.fields.name,
				nameWithoutRegistry: tt.fields.nameWithoutRegistry,
				registry:            tt.fields.registry,
				tag:                 tt.fields.tag,
				pullSecrets:         tt.fields.pullSecrets,
			}
			if got := i.GetNameWithoutRegistry(); got != tt.want {
				t.Errorf("GetNameWithoutRegistry() = %v, want %v", got, tt.want)
			}
		})
	}
}
