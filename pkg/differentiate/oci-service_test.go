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
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/fwiedmann/differ/pkg/registry"
)

var (
	initAPIClientFun = func(_ http.Client, _ registry.OciImage) OciRegistryAPIClient {
		return &ociAPIMock
	}
	ociServiceTestImages = []Image{
		{
			ID:       "1",
			Registry: "docker.com",
			Name:     "tomcat",
			Tag:      "1.8",
			Auth:     nil,
		},
		{
			ID:       "2",
			Registry: "docker.com",
			Name:     "tomcat",
			Tag:      "2.0",
			Auth:     nil,
		},
		{
			ID:       "3",
			Registry: "gitlab.com",
			Name:     "differ",
			Tag:      "1.0.0",
			Auth:     nil,
		},

		{
			ID:       "4",
			Registry: "github.com",
			Name:     "health",
			Tag:      "1.0.2",
			Auth: []*PullSecret{
				{
					Username: "admin",
					Password: "admin",
				},
			},
		},
	}
)

func TestNewOCIRegistryService(t *testing.T) {
	apiClient := ociAPIClientMOCK{}
	rp := repositoryMock{}
	workerCtx, cancel := context.WithCancel(context.TODO())
	initAPIClientFun := func(_ http.Client, _ registry.OciImage) OciRegistryAPIClient {
		return &apiClient
	}
	type args struct {
		ctx                 context.Context
		rp                  Repository
		initOCIAPIClientFun func(c http.Client, img registry.OciImage) OciRegistryAPIClient
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ValidNewOCIRegistryService",
			args: args{
				ctx:                 workerCtx,
				rp:                  rp,
				initOCIAPIClientFun: initAPIClientFun,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewOCIRegistryService(tt.args.ctx, tt.args.rp, tt.args.initOCIAPIClientFun)
			_, ok := svc.(*OCIRegistryService)
			if !ok {
				t.Errorf("NewOCIRegistryService() = returned service is not the type of OCIRegistryService")
			}
			cancel()

		})
	}
}

func TestOCIRegistryService_AddImage(t *testing.T) {

	rp := repositoryMock{}
	workerCtx, cancel := context.WithCancel(context.TODO())

	type fields struct {
		rp                  Repository
		workers             map[string]*Worker
		notifiers           []chan<- NotificationEvent
		workerNotification  chan NotificationEvent
		workerMtx           sync.Mutex
		workerCtx           context.Context
		initOCIAPIClientFun func(c http.Client, img registry.OciImage) OciRegistryAPIClient
	}
	type args struct {
		ctx    context.Context
		images []Image
	}

	type want struct {
		err         bool
		workerCount int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "IncreaseWorkerCount",
			fields: fields{
				rp:                  rp,
				workers:             make(map[string]*Worker),
				notifiers:           make([]chan<- NotificationEvent, 0),
				workerNotification:  make(chan NotificationEvent),
				workerMtx:           sync.Mutex{},
				workerCtx:           workerCtx,
				initOCIAPIClientFun: initAPIClientFun,
			},
			args: args{
				ctx: workerCtx,
				images: []Image{
					{
						Registry: "docker.com",
						Name:     "differ",
						Tag:      "1.0.0",
						Auth:     nil,
					},
					{
						Registry: "docker.com",
						Name:     "differ",
						Tag:      "1.1.0",
						Auth:     nil,
					},
					{
						Registry: "docker.com",
						Name:     "tomcat",
						Tag:      "1.0.0",
						Auth:     nil,
					},
					{
						Registry: "gitlab.com",
						Name:     "differ",
						Tag:      "1.0.0",
						Auth:     nil,
					},
				},
			},
			want: want{
				err:         false,
				workerCount: 3,
			},
		},
		{
			name: "RepositoryError",
			fields: fields{
				rp: repositoryMock{
					images: nil,
					addErr: fmt.Errorf("error"),
				},
				workers:             make(map[string]*Worker),
				notifiers:           make([]chan<- NotificationEvent, 0),
				workerNotification:  make(chan NotificationEvent),
				workerMtx:           sync.Mutex{},
				workerCtx:           workerCtx,
				initOCIAPIClientFun: initAPIClientFun,
			},
			args: args{
				ctx: workerCtx,
				images: []Image{
					{
						Registry: "docker.com",
						Name:     "differ",
						Tag:      "1.0.0",
						Auth:     nil,
					},
				},
			},
			want: want{
				err:         true,
				workerCount: 0,
			},
		},
	}
	defer cancel()
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			O := &OCIRegistryService{
				rp:                  tt.fields.rp,
				workers:             tt.fields.workers,
				notifiers:           tt.fields.notifiers,
				workerNotification:  tt.fields.workerNotification,
				workerMtx:           tt.fields.workerMtx,
				workerCtx:           tt.fields.workerCtx,
				initOCIAPIClientFun: tt.fields.initOCIAPIClientFun,
			}
			for _, i := range tt.args.images {
				if err := O.AddImage(tt.args.ctx, i); (err != nil) != tt.want.err {
					t.Errorf("AddImage() error = %v, want %v", err, tt.want)
				}
			}
			if tt.want.workerCount != len(O.workers) {
				t.Errorf("AddImage() Missmatch of desired worker count = want %d, got %d", tt.want.workerCount, len(O.workers))
			}

		})
	}
}

func TestOCIRegistryService_DeleteImage(t *testing.T) {
	type fields struct {
		rp                  Repository
		workers             map[string]*Worker
		notifiers           []chan<- NotificationEvent
		workerNotification  chan NotificationEvent
		workerMtx           sync.Mutex
		workerCtx           context.Context
		initOCIAPIClientFun func(c http.Client, img registry.OciImage) OciRegistryAPIClient
	}
	type args struct {
		ctx   context.Context
		image Image
	}

	type want struct {
		err         bool
		workerCount int
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "RemainWorkerCount",
			fields: fields{
				workers:             make(map[string]*Worker),
				notifiers:           make([]chan<- NotificationEvent, 0),
				workerNotification:  make(chan NotificationEvent),
				workerMtx:           sync.Mutex{},
				initOCIAPIClientFun: initAPIClientFun,
				rp:                  repositoryMock{images: ociServiceTestImages},
			},
			args: args{
				ctx: context.TODO(),
				image: Image{
					ID:       "1",
					Registry: "docker.com",
					Name:     "tomcat",
					Tag:      "1.8",
					Auth:     nil,
				},
			},
			want: want{
				err:         false,
				workerCount: 3,
			},
		},
		{
			name: "DecreaseWorkerCount",
			fields: fields{
				workers:             make(map[string]*Worker),
				notifiers:           make([]chan<- NotificationEvent, 0),
				workerNotification:  make(chan NotificationEvent),
				workerMtx:           sync.Mutex{},
				initOCIAPIClientFun: initAPIClientFun,
				rp:                  repositoryMock{},
			},
			args: args{
				ctx: context.TODO(),
				image: Image{
					ID:       "1",
					Registry: "docker.com",
					Name:     "tomcat",
					Tag:      "1.8",
					Auth:     nil,
				},
			},
			want: want{
				err:         false,
				workerCount: 2,
			},
		},
		{
			name: "RepositoryDeleteError",
			fields: fields{
				workers:             make(map[string]*Worker),
				notifiers:           make([]chan<- NotificationEvent, 0),
				workerNotification:  make(chan NotificationEvent),
				workerMtx:           sync.Mutex{},
				initOCIAPIClientFun: initAPIClientFun,
				rp:                  repositoryMock{deleteErr: fmt.Errorf("error"), images: ociServiceTestImages},
			},
			args: args{
				ctx: context.TODO(),
				image: Image{
					ID:       "1",
					Registry: "docker.com",
					Name:     "tomcat",
					Tag:      "1.8",
					Auth:     nil,
				},
			},
			want: want{
				err:         true,
				workerCount: 3,
			},
		},
		{
			name: "RepositoryListError",
			fields: fields{
				workers:             make(map[string]*Worker),
				notifiers:           make([]chan<- NotificationEvent, 0),
				workerNotification:  make(chan NotificationEvent),
				workerMtx:           sync.Mutex{},
				initOCIAPIClientFun: initAPIClientFun,
				rp:                  repositoryMock{listErr: fmt.Errorf("error"), images: ociServiceTestImages},
			},
			args: args{
				ctx: context.TODO(),
				image: Image{
					ID:       "1",
					Registry: "docker.com",
					Name:     "tomcat",
					Tag:      "1.8",
					Auth:     nil,
				},
			},
			want: want{
				err:         true,
				workerCount: 3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			O := &OCIRegistryService{
				rp:                  tt.fields.rp,
				workers:             tt.fields.workers,
				notifiers:           tt.fields.notifiers,
				workerNotification:  tt.fields.workerNotification,
				workerMtx:           tt.fields.workerMtx,
				workerCtx:           ctx,
				initOCIAPIClientFun: tt.fields.initOCIAPIClientFun,
			}

			for _, i := range ociServiceTestImages {
				if err := O.AddImage(context.TODO(), i); err != nil {
					t.Fatalf("AddImage() error = %v", err)
				}
			}

			if err := O.DeleteImage(tt.args.ctx, tt.args.image); (err != nil) != tt.want.err {
				t.Errorf("DeleteImage() error = %v, want %v", err, tt.want)
			}

			if tt.want.workerCount != len(O.workers) {
				t.Errorf("DeleteImage() Missmatch of desired worker count = want %d, got %d", tt.want.workerCount, len(O.workers))
			}

			for _, w := range O.workers {
				go w.Stop()
			}

			cancel()

		})
	}
}

func TestOCIRegistryService_UpdateImage(t *testing.T) {
	type fields struct {
		rp                  Repository
		workers             map[string]*Worker
		notifiers           []chan<- NotificationEvent
		workerNotification  chan NotificationEvent
		workerMtx           sync.Mutex
		workerCtx           context.Context
		initOCIAPIClientFun func(c http.Client, img registry.OciImage) OciRegistryAPIClient
	}
	type args struct {
		ctx   context.Context
		image Image
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{

		{
			name: "NoRepositoryError",
			fields: fields{
				rp: repositoryMock{},
			},
			args: args{
				ctx:   context.TODO(),
				image: Image{},
			},
			wantErr: false,
		},
		{
			name: "RepositoryError",
			fields: fields{
				rp: repositoryMock{updateErr: fmt.Errorf("error")},
			},
			args: args{
				ctx:   context.TODO(),
				image: Image{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			O := &OCIRegistryService{
				rp:                  tt.fields.rp,
				workers:             tt.fields.workers,
				notifiers:           tt.fields.notifiers,
				workerNotification:  tt.fields.workerNotification,
				workerMtx:           tt.fields.workerMtx,
				workerCtx:           tt.fields.workerCtx,
				initOCIAPIClientFun: tt.fields.initOCIAPIClientFun,
			}
			if err := O.UpdateImage(tt.args.ctx, tt.args.image); (err != nil) != tt.wantErr {
				t.Errorf("UpdateImage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOCIRegistryService_ListImages(t *testing.T) {
	type fields struct {
		rp                  Repository
		workers             map[string]*Worker
		notifiers           []chan<- NotificationEvent
		workerNotification  chan NotificationEvent
		workerMtx           sync.Mutex
		workerCtx           context.Context
		initOCIAPIClientFun func(c http.Client, img registry.OciImage) OciRegistryAPIClient
	}
	type args struct {
		ctx  context.Context
		opts ListOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Image
		wantErr bool
	}{
		{
			name: "NoRepositoryError",
			fields: fields{
				rp: repositoryMock{},
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
		},
		{
			name: "RepositoryError",
			fields: fields{
				rp: repositoryMock{listErr: fmt.Errorf("error")},
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			O := &OCIRegistryService{
				rp:                  tt.fields.rp,
				workers:             tt.fields.workers,
				notifiers:           tt.fields.notifiers,
				workerNotification:  tt.fields.workerNotification,
				workerMtx:           tt.fields.workerMtx,
				workerCtx:           tt.fields.workerCtx,
				initOCIAPIClientFun: tt.fields.initOCIAPIClientFun,
			}
			got, err := O.ListImages(tt.args.ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListImages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListImages() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOCIRegistryService_Notify(t *testing.T) {

	initAPIClientFun := func(_ http.Client, _ registry.OciImage) OciRegistryAPIClient {
		return &ociAPIClientMOCK{}
	}

	type fields struct {
		rp                  Repository
		workerCtx           context.Context
		initOCIAPIClientFun func(c http.Client, img registry.OciImage) OciRegistryAPIClient
	}
	type args struct {
		event chan NotificationEvent
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "",
			fields: fields{
				rp:                  repositoryMock{},
				initOCIAPIClientFun: initAPIClientFun,
			},
			args: args{
				event: make(chan NotificationEvent),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			ociService := NewOCIRegistryService(ctx, tt.fields.rp, tt.fields.initOCIAPIClientFun)
			ociService.Notify(tt.args.event)

			val, ok := ociService.(*OCIRegistryService)
			if !ok {
				t.Fatalf("Notify(): could not do type assertion of %+v, to *OCIRegistryService", ociService)
			}

			go func() {
				val.workerNotification <- NotificationEvent{}
			}()

			select {
			case <-time.After(time.Second * 5):
				cancel()
				t.Fatal("Notify(): did not receive desired NotificationEvent.")
			case <-tt.args.event:
				cancel()
				return
			}

		})

	}
}
