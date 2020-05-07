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

package registry

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

const (
	testRealm        = "https://felixwiedmann.de"
	testRealmService = "container_registry"
)

type readCloser struct{}

func (readCloser) Read(_ []byte) (n int, err error) {
	return n, nil
}

func (readCloser) Close() error {
	return nil
}

var (
	validRealmRequest = func(request *http.Request) (*http.Response, error) {
		h := http.Header{}
		h.Add("WWW-Authenticate", "Bearer realm=\""+testRealm+"\",service=\""+testRealmService+"\"")
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     h,
			Body:       readCloser{},
		}, nil
	}

	invalidRequestError = func(request *http.Request) (*http.Response, error) {
		return &http.Response{}, fmt.Errorf("api error")
	}

	invalidRealmRequestStatusCode = func(request *http.Request) (*http.Response, error) {
		h := http.Header{}
		h.Add("WWW-Authenticate", "Bearer realm=\"https://gitlab.com/jwt/auth\",service=\"container_registry\"")
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     h,
			Body:       readCloser{},
		}, nil
	}

	invalidRealmRequestHeaderNotFound = func(request *http.Request) (*http.Response, error) {
		h := http.Header{}
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     h,
			Body:       readCloser{},
		}, nil
	}

	invalidRealmRequestHeaderNoRealmURL = func(request *http.Request) (*http.Response, error) {
		h := http.Header{}
		h.Add("WWW-Authenticate", "service=\"container_registry\"")
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     h,
			Body:       readCloser{},
		}, nil
	}

	invalidRealmRequestHeaderNoService = func(request *http.Request) (*http.Response, error) {
		h := http.Header{}
		h.Add("WWW-Authenticate", "Bearer realm=\"https://gitlab.com/jwt/auth\"")
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     h,
			Body:       readCloser{},
		}, nil
	}

	invalidTokenRequestStatusUnauthorized = func(request *http.Request) (*http.Response, error) {
		h := http.Header{}
		h.Add("WWW-Authenticate", "Bearer realm=\"https://gitlab.com/jwt/auth\",service=\"container_registry\"")
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     h,
			Body:       readCloser{},
			Request: &http.Request{
				URL: &url.URL{Host: testRealmService},
			},
		}, nil
	}

	invalidTokenRequestStatusForbidden = func(request *http.Request) (*http.Response, error) {
		h := http.Header{}
		h.Add("WWW-Authenticate", "Bearer realm=\"https://gitlab.com/jwt/auth\",service=\"container_registry\"")
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Header:     h,
			Body:       readCloser{},
			Request: &http.Request{
				URL: &url.URL{Host: testRealmService},
			},
		}, nil
	}

	invalidTokenRequestStatusNotFound = func(request *http.Request) (*http.Response, error) {
		h := http.Header{}
		h.Add("WWW-Authenticate", "Bearer realm=\"https://gitlab.com/jwt/auth\",service=\"container_registry\"")
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Header:     h,
			Body:       readCloser{},
			Request: &http.Request{
				URL: &url.URL{Host: testRealmService},
			},
		}, nil
	}
)

type image struct {
	withoutRegistry string
	registryURL     string
}

func (i image) GetNameWithoutRegistry() string {
	return i.withoutRegistry
}

func (i image) GetRegistryURL() string {
	return i.registryURL
}

type roundTripper struct {
	requests map[string]func(request *http.Request) (*http.Response, error)
}

func (rt roundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	f, ok := rt.requests[fmt.Sprintf("%s%s", request.URL.Host, request.URL.Path)]
	if !ok {
		f, ok = rt.requests[fmt.Sprintf("https://%s?%s", request.URL.Host, request.URL.RawQuery)]
		if !ok {
			panic("OCI registry test suite: round tripper could not find a valid method")
		}
		return f(request)
	}
	return f(request)
}

func TestOciAPIClient_GetTagsForImage(t *testing.T) {

	type fields struct {
		Image       OciImage
		bearerToken string
		Client      http.Client
	}
	type args struct {
		ctx    context.Context
		secret OciPullSecret
	}

	image := image{
		withoutRegistry: "differ",
		registryURL:     "docker.com",
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		/*{
			name: "WithoutAuth",
			fields: fields{
				Image:       image,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2", image.registryURL): validRealmRequest,
						},
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				secret: nil,
			},
			want:    nil,
			wantErr: false,
		},*/
		{
			name: "invalidRequestRealmError",
			fields: fields{
				Image:       image,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2", image.registryURL): invalidRequestError,
						},
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				secret: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalidRealmRequestStatusCode",
			fields: fields{
				Image:       image,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2", image.registryURL): invalidRealmRequestStatusCode,
						},
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				secret: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalidRealmRequestHeaderNotFound",
			fields: fields{
				Image:       image,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2", image.registryURL): invalidRealmRequestHeaderNotFound,
						},
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				secret: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalidRealmRequestHeaderNoRealmURL",
			fields: fields{
				Image:       image,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2", image.registryURL): invalidRealmRequestHeaderNoRealmURL,
						},
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				secret: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalidRealmRequestHeaderNoService",
			fields: fields{
				Image:       image,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2", image.registryURL): invalidRealmRequestHeaderNoService,
						},
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				secret: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalidTokenRequestStatusUnauthorized",
			fields: fields{
				Image:       image,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2", image.registryURL): validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, image.withoutRegistry): invalidTokenRequestStatusUnauthorized,
						},
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				secret: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalidTokenRequestStatusForbidden",
			fields: fields{
				Image:       image,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2", image.registryURL): validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, image.withoutRegistry): invalidTokenRequestStatusForbidden,
						},
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				secret: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalidTokenRequestStatusNotFound",
			fields: fields{
				Image:       image,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2", image.registryURL): validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, image.withoutRegistry): invalidTokenRequestStatusNotFound,
						},
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				secret: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalidRequestTokenError",
			fields: fields{
				Image:       image,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2", image.registryURL): validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, image.withoutRegistry): invalidRequestError,
						},
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				secret: nil,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &OciAPIClient{
				Image:       tt.fields.Image,
				bearerToken: tt.fields.bearerToken,
				Client:      tt.fields.Client,
			}

			got, err := c.GetTagsForImage(tt.args.ctx, tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTagsForImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTagsForImage() got = %v, want %v", got, tt.want)
			}
		})
	}
}
