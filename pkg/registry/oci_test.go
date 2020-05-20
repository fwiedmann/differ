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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

const (
	testRealm        = "https://felixwiedmann.de"
	testRealmService = "container_registry"
	testBearerToken  = "6c2a7e12fe2c44e4a2707d9f1456ab2b"
)

type pullSecret struct {
	username, password string
}

func (ps *pullSecret) GetUsername() string {
	return ps.username
}

func (ps *pullSecret) GetPassword() string {
	return ps.password
}

type readCloser struct{}

func (readCloser) Read(_ []byte) (n int, err error) {
	return n, nil
}

func (readCloser) Close() error {
	return nil
}

var (
	testTagList       = []string{"1.0.0", "2.5.7", "8.9.1"}
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
	invalidTokenRequestEmptyToken = func(request *http.Request) (*http.Response, error) {
		h := http.Header{}

		tokenResponse, err := json.Marshal(&bearerToken{
			Token: "",
		})

		if err != nil {
			panic(err)
		}

		buf := &bytes.Buffer{}
		buf.Write(tokenResponse)

		h.Add("WWW-Authenticate", "Bearer realm=\"https://gitlab.com/jwt/auth\",service=\"container_registry\"")
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     h,
			Body:       ioutil.NopCloser(buf),
			Request: &http.Request{
				URL: &url.URL{Host: testRealmService},
			},
		}, nil
	}
	validTokenRequest = func(request *http.Request) (*http.Response, error) {
		h := http.Header{}

		tokenResponse, err := json.Marshal(&bearerToken{
			Token: testBearerToken,
		})

		if err != nil {
			panic(err)
		}

		buf := &bytes.Buffer{}
		buf.Write(tokenResponse)

		h.Add("WWW-Authenticate", "Bearer realm=\"https://gitlab.com/jwt/auth\",service=\"container_registry\"")
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     h,
			Body:       ioutil.NopCloser(buf),
			Request: &http.Request{
				URL: &url.URL{Host: testRealmService},
			},
		}, nil
	}
	invalidTagRequestError = func(request *http.Request) (*http.Response, error) {
		return &http.Response{}, fmt.Errorf("api error")
	}

	invalidTagRequestStatusUnauthorized = func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Authorization") != fmt.Sprintf("Bearer "+testBearerToken) {
			return nil, fmt.Errorf("authorization header is not valid")
		}
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     http.Header{},
			Body:       readCloser{},
			Request: &http.Request{
				URL: &url.URL{Host: testRealmService},
			},
		}, nil
	}

	invalidTagRequestStatusForbidden = func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Authorization") != fmt.Sprintf("Bearer "+testBearerToken) {
			return nil, fmt.Errorf("authorization header is not valid")
		}
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Header:     http.Header{},
			Body:       readCloser{},
			Request: &http.Request{
				URL: &url.URL{Host: testRealmService},
			},
		}, nil
	}

	invalidTagRequestStatusNotFound = func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Authorization") != fmt.Sprintf("Bearer "+testBearerToken) {
			return nil, fmt.Errorf("authorization header is not valid")
		}
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Header:     http.Header{},
			Body:       readCloser{},
			Request: &http.Request{
				URL: &url.URL{Host: testRealmService},
			},
		}, nil
	}

	validTagRequest = func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Authorization") != fmt.Sprintf("Bearer "+testBearerToken) {
			return nil, fmt.Errorf("authorization header is not valid")
		}

		tokenResponse, err := json.Marshal(&tagList{
			Tags: testTagList,
		})

		if err != nil {
			panic(err)
		}

		buf := &bytes.Buffer{}
		buf.Write(tokenResponse)

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
			Body:       ioutil.NopCloser(buf),
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

	imageWithoutAuth := image{
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
		{
			name: "WithoutAuth",
			fields: fields{
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL):                                                                  validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, imageWithoutAuth.withoutRegistry): validTokenRequest,
							fmt.Sprintf("%s/%s/%s/tags/list", imageWithoutAuth.registryURL, "v2", imageWithoutAuth.withoutRegistry):              validTagRequest,
						},
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				secret: nil,
			},
			want:    testTagList,
			wantErr: false,
		},
		{
			name: "WithAuth",
			fields: fields{
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL):                                                                  validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, imageWithoutAuth.withoutRegistry): validTokenRequest,
							fmt.Sprintf("%s/%s/%s/tags/list", imageWithoutAuth.registryURL, "v2", imageWithoutAuth.withoutRegistry):              validTagRequest,
						},
					},
				},
			},
			args: args{
				ctx: context.TODO(),
				secret: &pullSecret{
					username: "admin",
					password: "admin",
				},
			},
			want:    testTagList,
			wantErr: false,
		},
		{
			name: "invalidRequestRealmError",
			fields: fields{
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL): invalidRequestError,
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
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL): invalidRealmRequestStatusCode,
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
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL): invalidRealmRequestHeaderNotFound,
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
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL): invalidRealmRequestHeaderNoRealmURL,
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
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL): invalidRealmRequestHeaderNoService,
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
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL):                                                                  validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, imageWithoutAuth.withoutRegistry): invalidTokenRequestStatusUnauthorized,
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
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL):                                                                  validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, imageWithoutAuth.withoutRegistry): invalidTokenRequestStatusForbidden,
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
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL):                                                                  validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, imageWithoutAuth.withoutRegistry): invalidTokenRequestStatusNotFound,
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
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL):                                                                  validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, imageWithoutAuth.withoutRegistry): invalidRequestError,
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
			name: "invalidTokenRequestEmptyToken",
			fields: fields{
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL):                                                                  validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, imageWithoutAuth.withoutRegistry): invalidTokenRequestEmptyToken,
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
			name: "invalidTagRequestError",
			fields: fields{
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL):                                                                  validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, imageWithoutAuth.withoutRegistry): validTokenRequest,
							fmt.Sprintf("%s/%s/%s/tags/list", imageWithoutAuth.registryURL, "v2", imageWithoutAuth.withoutRegistry):              invalidTagRequestError,
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
			name: "invalidTagRequestStatusUnauthorized",
			fields: fields{
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL):                                                                  validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, imageWithoutAuth.withoutRegistry): validTokenRequest,
							fmt.Sprintf("%s/%s/%s/tags/list", imageWithoutAuth.registryURL, "v2", imageWithoutAuth.withoutRegistry):              invalidTagRequestStatusUnauthorized,
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
			name: "invalidTagRequestStatusForbidden",
			fields: fields{
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL):                                                                  validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, imageWithoutAuth.withoutRegistry): validTokenRequest,
							fmt.Sprintf("%s/%s/%s/tags/list", imageWithoutAuth.registryURL, "v2", imageWithoutAuth.withoutRegistry):              invalidTagRequestStatusForbidden,
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
			name: "invalidTagRequestStatusNotFound",
			fields: fields{
				Image:       imageWithoutAuth,
				bearerToken: "",
				Client: http.Client{
					Transport: roundTripper{
						map[string]func(request *http.Request) (*http.Response, error){
							fmt.Sprintf("%s/v2/", imageWithoutAuth.registryURL):                                                                  validRealmRequest,
							fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", testRealm, testRealmService, imageWithoutAuth.withoutRegistry): validTokenRequest,
							fmt.Sprintf("%s/%s/%s/tags/list", imageWithoutAuth.registryURL, "v2", imageWithoutAuth.withoutRegistry):              invalidTagRequestStatusNotFound,
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
