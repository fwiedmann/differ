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

package differentiate

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"go.uber.org/ratelimit"
)

type repositoryMock struct {
	images []Image
	err    error
}

func (r repositoryMock) ListImages(_ context.Context, _ ListOptions) ([]Image, error) {
	return r.images, r.err
}

type ociAPIClientMOCK struct {
	tags []string
	err  error
}

func (o ociAPIClientMOCK) GetTagsForImage(_ context.Context, _ OciPullSecret) ([]string, error) {
	return o.tags, o.err
}

var (
	imageRemoteTags = []string{"1.0.0,2.0.0,3.0.0"}
	image           = Image{
		ID:       "1111",
		Registry: "differ.com",
		Name:     "differ",
		Tag:      "1.0.0",
		Auth:     nil,
	}

	rpMock = repositoryMock{images: []Image{image}}

	ociAPIMock = ociAPIClientMOCK{
		tags: imageRemoteTags,
		err:  nil,
	}
	rl       = ratelimit.New(5)
	infoChan = make(chan NotificationEvent)
)

func TestStartNewImageWorker(t *testing.T) {

	type args struct {
		ctx         context.Context
		client      OciRegistryAPIClient
		registry    string
		imageName   string
		rateLimiter ratelimit.Limiter
		info        chan<- NotificationEvent
		repository  ListImagesRepository
	}
	tests := []struct {
		name string
		args args
		want *Worker
	}{
		{
			name: "ValidStartNewImageWorkerFun",
			args: args{
				ctx:         context.TODO(),
				client:      ociAPIMock,
				registry:    image.Registry,
				imageName:   image.Name,
				rateLimiter: rl,
				info:        infoChan,
				repository:  rpMock,
			},
			want: &Worker{
				imageName:   image.Name,
				registry:    image.Registry,
				rp:          rpMock,
				mutex:       sync.RWMutex{},
				informChan:  infoChan,
				rateLimiter: rl,
				client:      ociAPIMock,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StartNewImageWorker(tt.args.ctx, tt.args.client, tt.args.registry, tt.args.imageName, tt.args.rateLimiter, tt.args.info, tt.args.repository); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StartNewImageWorker() = %+v, want %+v", got, tt.want)
			}
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
				imageName:   image.Name,
				registry:    image.Registry,
				rp:          rpMock,
				mutex:       sync.RWMutex{},
				informChan:  infoChan,
				rateLimiter: rl,
				client:      ociAPIMock,
				stop:        make(chan struct{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Worker{
				imageName:   tt.fields.imageName,
				registry:    tt.fields.registry,
				rp:          tt.fields.rp,
				mutex:       tt.fields.mutex,
				informChan:  tt.fields.informChan,
				rateLimiter: tt.fields.rateLimiter,
				client:      tt.fields.client,
				stop:        tt.fields.stop,
			}
			w.Stop()

			select {
			case <-time.After(time.Second * 5):
				t.Errorf("Worker did not send stop signal")
			case <-w.stop:
				return
			}
		})
	}
}
