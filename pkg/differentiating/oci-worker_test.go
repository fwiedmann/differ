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

package differentiating

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/fwiedmann/differ/pkg/registry"
	"go.uber.org/ratelimit"
)

type repositoryMock struct {
	images    []Image
	addErr    error
	deleteErr error
	updateErr error
	listErr   error
}

func (r repositoryMock) AddImage(_ context.Context, _ Image) error {
	return r.addErr
}

func (r repositoryMock) DeleteImage(_ context.Context, _ Image) error {
	return r.deleteErr
}

func (r repositoryMock) UpdateImage(_ context.Context, _ Image) error {
	return r.updateErr
}

func (r repositoryMock) ListImages(_ context.Context, _ ListOptions) ([]Image, error) {
	return r.images, r.listErr
}

type ociAPIClientMOCK struct {
	tags []string
	err  error
}

func (o ociAPIClientMOCK) GetTagsForImage(_ context.Context, _ registry.OciPullSecret) ([]string, error) {
	return o.tags, o.err
}

var (
	imageRemoteTags  = []string{"1.0.0", "2.0.0", "3.0.0"}
	imageWithoutAuth = Image{
		ID:       "1111",
		Registry: "differ.com",
		Name:     "differ",
		Tag:      "1.0.0",
		Auth:     nil,
	}

	imageWithAuth = Image{
		ID:       "1111",
		Registry: "differ.com",
		Name:     "differ",
		Tag:      "1.0.0",
		Auth: []*PullSecret{{
			Username: "admin",
			Password: "admin"},
		},
	}

	rpWithoutAuthMockMock = repositoryMock{images: []Image{imageWithoutAuth}}
	rpWithAuthMockMock    = repositoryMock{images: []Image{imageWithAuth}}

	ociAPIMock = ociAPIClientMOCK{
		tags: imageRemoteTags,
		err:  nil,
	}

	ociAPIErrorMock = ociAPIClientMOCK{
		tags: imageRemoteTags,
		err:  fmt.Errorf("error"),
	}
	rl       = ratelimit.New(5)
	infoChan = make(chan NotificationEvent)
)

func TestStartNewImageWorker(t *testing.T) {

	type args struct {
		client      OciRegistryAPIClient
		registry    string
		imageName   string
		rateLimiter ratelimit.Limiter
		info        chan NotificationEvent
		repository  ListImagesRepository
	}
	tests := []struct {
		name string
		args args
		want *Worker
	}{
		{
			name: "ImageWithoutAuth",
			args: args{
				client:      ociAPIMock,
				registry:    imageWithoutAuth.Registry,
				imageName:   imageWithoutAuth.Name,
				rateLimiter: rl,
				info:        infoChan,
				repository:  rpWithoutAuthMockMock,
			},
			want: &Worker{
				imageName:   imageWithoutAuth.Name,
				registry:    imageWithoutAuth.Registry,
				rp:          rpWithoutAuthMockMock,
				mutex:       sync.RWMutex{},
				informChan:  infoChan,
				rateLimiter: rl,
				client:      ociAPIMock,
			},
		},
		{
			name: "ImageWithoutAuthAPIError",
			args: args{
				client:      ociAPIErrorMock,
				registry:    imageWithoutAuth.Registry,
				imageName:   imageWithoutAuth.Name,
				rateLimiter: rl,
				info:        infoChan,
				repository:  rpWithoutAuthMockMock,
			},
			want: &Worker{
				imageName:   imageWithoutAuth.Name,
				registry:    imageWithoutAuth.Registry,
				rp:          rpWithoutAuthMockMock,
				mutex:       sync.RWMutex{},
				informChan:  infoChan,
				rateLimiter: rl,
				client:      ociAPIErrorMock,
			},
		},
		{
			name: "ImageWithAuth",
			args: args{
				client:      ociAPIMock,
				registry:    imageWithAuth.Registry,
				imageName:   imageWithAuth.Name,
				rateLimiter: rl,
				info:        infoChan,
				repository:  rpWithAuthMockMock,
			},
			want: &Worker{
				imageName:   imageWithAuth.Name,
				registry:    imageWithAuth.Registry,
				rp:          rpWithAuthMockMock,
				mutex:       sync.RWMutex{},
				informChan:  infoChan,
				rateLimiter: rl,
				client:      ociAPIMock,
			},
		},
		{
			name: "ImageWithAuthAPIError",
			args: args{
				client:      ociAPIErrorMock,
				registry:    imageWithAuth.Registry,
				imageName:   imageWithAuth.Name,
				rateLimiter: rl,
				info:        infoChan,
				repository:  rpWithAuthMockMock,
			},
			want: &Worker{
				imageName:   imageWithAuth.Name,
				registry:    imageWithAuth.Registry,
				rp:          rpWithAuthMockMock,
				mutex:       sync.RWMutex{},
				informChan:  infoChan,
				rateLimiter: rl,
				client:      ociAPIErrorMock,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			worker := StartNewImageWorker(ctx, tt.args.client, tt.args.registry, tt.args.imageName, tt.args.rateLimiter, tt.args.info, tt.args.repository)
			tt.want.stop = worker.stop
			if !reflect.DeepEqual(worker, tt.want) {
				t.Errorf("StartNewImageWorker() = %+v, want %+v", worker, tt.want)
			}
			<-tt.args.info
			cancel()
		})
	}
}

func TestWorker_Stop(t *testing.T) {
	type fields struct {
		imageName   string
		registry    string
		rp          ListImagesRepository
		mutex       sync.RWMutex
		informChan  chan<- NotificationEvent
		rateLimiter ratelimit.Limiter
		client      OciRegistryAPIClient
		stop        chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "ValidWorkerStop",
			fields: fields{
				imageName:   imageWithoutAuth.Name,
				registry:    imageWithoutAuth.Registry,
				rp:          rpWithoutAuthMockMock,
				mutex:       sync.RWMutex{},
				informChan:  infoChan,
				rateLimiter: rl,
				client:      ociAPIMock,
				stop:        make(chan struct{}),
			},
		},
	}
	for _, tt := range tests { //nolint
		t.Run(tt.name, func(t *testing.T) {
			w := &Worker{
				imageName:   tt.fields.imageName,
				registry:    tt.fields.registry,
				rp:          tt.fields.rp,
				mutex:       tt.fields.mutex, //nolint
				informChan:  tt.fields.informChan,
				rateLimiter: tt.fields.rateLimiter,
				client:      tt.fields.client,
				stop:        tt.fields.stop,
			}
			go w.Stop()

			select {
			case <-time.After(time.Second * 5):
				t.Errorf("Worker did not send stop signal")
			case <-w.stop:
				return
			}
		})
	}
}
