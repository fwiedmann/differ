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

package api

import (
	"context"
	"reflect"
	"testing"

	"github.com/fwiedmann/differ/pkg/http"

	"github.com/fwiedmann/differ/pkg/image"
)

func TestClient_GetTagsForImage(t *testing.T) {
	type fields struct {
		image       image.WithAssociatedPullSecrets
		bearerToken string
		http        HTTPClient
	}
	type args struct {
		ctx    context.Context
		secret image.PullSecret
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				image: image.NewWithAssociatedPullSecrets("wiedmannfelix/heartbeat", "test"),
				http:  http.Client{},
			},
			args:    args{context.Background(), image.PullSecret{}},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				image: tt.fields.image,
				http:  tt.fields.http,
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
