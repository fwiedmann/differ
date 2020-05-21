/*
 * MIT License
 *
 * Copyright (ctx) 2019 Felix Wiedmann
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

package memory

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/fwiedmann/differ/pkg/differentiating"
)

func TestNewMemoryStorage(t *testing.T) {
	tests := []struct {
		name string
		want *Storage
	}{
		{
			name: "Valid",
			want: &Storage{
				mtx:    sync.RWMutex{},
				images: make(map[string]differentiating.Image),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemoryStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemoryStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_AddImage(t *testing.T) {
	type fields struct {
		mtx    sync.RWMutex
		images map[string]differentiating.Image
	}
	type args struct {
		ctx context.Context
		img differentiating.Image
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Valid",
			fields: fields{
				mtx:    sync.RWMutex{},
				images: make(map[string]differentiating.Image),
			},
			args: args{
				ctx: context.TODO(),
				img: differentiating.Image{
					ID:       "1",
					Registry: "docker.com",
					Name:     "wiedmannfelix/differ",
					Tag:      "187",
					Auth:     nil,
				},
			},
			wantErr: false,
		},
		{
			name: "NoIDError",
			fields: fields{
				mtx:    sync.RWMutex{},
				images: make(map[string]differentiating.Image),
			},
			args: args{
				ctx: context.TODO(),
				img: differentiating.Image{
					ID:       "",
					Registry: "docker.com",
					Name:     "wiedmannfelix/differ",
					Tag:      "187",
					Auth:     nil,
				},
			},
			wantErr: true,
		},
		{
			name: "ImageAlreadyExists",
			fields: fields{
				mtx:    sync.RWMutex{},
				images: map[string]differentiating.Image{"1": differentiating.Image{ID: "1"}},
			},
			args: args{
				ctx: context.TODO(),
				img: differentiating.Image{
					ID:       "1",
					Registry: "docker.com",
					Name:     "wiedmannfelix/differ",
					Tag:      "187",
					Auth:     nil,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests { //nolint
		t.Run(tt.name, func(t *testing.T) {
			s := Storage{
				mtx:    tt.fields.mtx, //nolint
				images: tt.fields.images,
			}
			err := s.AddImage(tt.args.ctx, tt.args.img)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddImage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if _, ok := s.images[tt.args.img.ID]; !ok && err == nil {
				t.Errorf("AddImage() error = image is not stored")
			}
		})
	}
}

func TestStorage_DeleteImage(t *testing.T) {
	type fields struct {
		mtx    sync.RWMutex
		images map[string]differentiating.Image
	}
	type args struct {
		ctx context.Context
		img differentiating.Image
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Valid",
			fields: fields{
				mtx:    sync.RWMutex{},
				images: map[string]differentiating.Image{"1": differentiating.Image{ID: "1"}},
			},
			args: args{
				ctx: context.TODO(),
				img: differentiating.Image{
					ID: "1",
				},
			},
			wantErr: false,
		},
		{
			name: "NoIDError",
			fields: fields{
				mtx:    sync.RWMutex{},
				images: map[string]differentiating.Image{"1": differentiating.Image{ID: "1"}},
			},
			args: args{
				ctx: context.TODO(),
				img: differentiating.Image{
					ID: "",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests { //nolint
		t.Run(tt.name, func(t *testing.T) {
			s := Storage{
				mtx:    tt.fields.mtx, //nolint
				images: tt.fields.images,
			}

			err := s.DeleteImage(tt.args.ctx, tt.args.img)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteImage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if _, ok := s.images[tt.args.img.ID]; ok && err == nil {
				t.Errorf("DeleteImage() error = image is present in map but should be deleted")
			}
		})
	}
}

func TestStorage_UpdateImage(t *testing.T) {
	type fields struct {
		mtx    sync.RWMutex
		images map[string]differentiating.Image
	}
	type args struct {
		ctx context.Context
		img differentiating.Image
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Valid",
			fields: fields{
				mtx:    sync.RWMutex{},
				images: map[string]differentiating.Image{"1": differentiating.Image{ID: "1"}},
			},
			args: args{
				ctx: context.TODO(),
				img: differentiating.Image{
					ID: "1",
				},
			},
			wantErr: false,
		},
		{
			name: "NoIDError",
			fields: fields{
				mtx:    sync.RWMutex{},
				images: map[string]differentiating.Image{"1": differentiating.Image{ID: "1"}},
			},
			args: args{
				ctx: context.TODO(),
				img: differentiating.Image{
					ID: "2",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests { //nolint
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				mtx:    tt.fields.mtx, //nolint
				images: tt.fields.images,
			}
			if err := s.UpdateImage(tt.args.ctx, tt.args.img); (err != nil) != tt.wantErr {
				t.Errorf("UpdateImage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorage_ListImages(t *testing.T) {

	images := map[string]differentiating.Image{
		"1": differentiating.Image{
			ID:       "1",
			Registry: "docker.com",
			Name:     "differ",
			Tag:      "1.0.0",
			Auth:     nil,
		},
		"2": differentiating.Image{
			ID:       "2",
			Registry: "github.com",
			Name:     "differ",
			Tag:      "1.0.0",
			Auth:     nil,
		},
		"3": differentiating.Image{
			ID:       "3",
			Registry: "github.com",
			Name:     "tomcat",
			Tag:      "1.0.0",
			Auth:     nil,
		},
	}
	type fields struct {
		mtx    sync.RWMutex
		images map[string]differentiating.Image
	}
	type args struct {
		ctx  context.Context
		opts differentiating.ListOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []differentiating.Image
		wantErr bool
	}{
		{
			name: "OnlyName",
			fields: fields{
				mtx:    sync.RWMutex{},
				images: images,
			},
			args: args{
				ctx: context.TODO(),
				opts: differentiating.ListOptions{
					ImageName: "differ",
				},
			},
			want: []differentiating.Image{
				{
					ID:       "1",
					Registry: "docker.com",
					Name:     "differ",
					Tag:      "1.0.0",
					Auth:     nil,
				},
				{
					ID:       "2",
					Registry: "github.com",
					Name:     "differ",
					Tag:      "1.0.0",
					Auth:     nil,
				},
			},
			wantErr: false,
		},
		{
			name: "OnlyRegistry",
			fields: fields{
				mtx:    sync.RWMutex{},
				images: images,
			},
			args: args{
				ctx: context.TODO(),
				opts: differentiating.ListOptions{
					Registry: "docker.com",
				},
			},
			want: []differentiating.Image{
				{
					ID:       "1",
					Registry: "docker.com",
					Name:     "differ",
					Tag:      "1.0.0",
					Auth:     nil,
				},
			},
			wantErr: false,
		},
		{
			name: "NameAndRegistry",
			fields: fields{
				mtx:    sync.RWMutex{},
				images: images,
			},
			args: args{
				ctx: context.TODO(),
				opts: differentiating.ListOptions{
					ImageName: "differ",
					Registry:  "github.com",
				},
			},
			want: []differentiating.Image{
				{
					ID:       "2",
					Registry: "github.com",
					Name:     "differ",
					Tag:      "1.0.0",
					Auth:     nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests { //nolint
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				mtx:    tt.fields.mtx, //nolint
				images: tt.fields.images,
			}
			got, err := s.ListImages(tt.args.ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListImages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, wantImage := range tt.want {
				var found bool
				for _, gotImage := range got {
					if reflect.DeepEqual(wantImage, gotImage) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ListImages() got = %v, want %v", got, tt.want)
				}
			}

		})
	}
}
