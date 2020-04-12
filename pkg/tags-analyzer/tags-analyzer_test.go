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

package tags_analyzer

import (
	"reflect"
	"regexp"
	"testing"
)

func TestGetRegexExprForTag(t *testing.T) {
	type args struct {
		tag string
	}
	tests := []struct {
		name    string
		args    args
		want    *regexp.Regexp
		wantErr bool
	}{
		{name: "Valid1", args: args{tag: "1.14.2-nanoserver-1809"}, want: regexp.MustCompile("^\\d+.\\d+.\\d+-nanoserver-\\d+$"), wantErr: false},
		{name: "Valid2", args: args{tag: "8.5-jdk14-openjdk-oracle"}, want: regexp.MustCompile("^\\d+.\\d+-jdk\\d+-openjdk-oracle$"), wantErr: false},
		{name: "Valid3", args: args{tag: "v2.17.0"}, want: regexp.MustCompile("^v\\d+.\\d+.\\d+$"), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRegexExprForTag(tt.args.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRegexExprForTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRegexExprForTag() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLatestTagWithRegexExpr(t *testing.T) {
	type args struct {
		tags []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "ValidPatchWithoutString", args: args{tags: []string{"1.1.1", "1.1.2", "1.1.3", "1.1.4", "1.1.5"}}, want: "1.1.5", wantErr: false},
		{name: "ValidPatchWithString", args: args{tags: []string{"v1.1.1", "v1.1.2", "v1.1.3", "v1.1.4", "v1.1.5"}}, want: "v1.1.5", wantErr: false},
		{name: "ValidMinorWithoutString", args: args{tags: []string{"1.5.1", "1.4.2", "1.5.3", "1.3.4", "1.1.5"}}, want: "1.5.3", wantErr: false},
		{name: "ValidMinorWithString", args: args{tags: []string{"v1.5.1", "v1.4.2", "v1.5.3", "v1.3.4", "v1.1.5"}}, want: "v1.5.3", wantErr: false},
		{name: "ValidMajorWithoutString", args: args{tags: []string{"2.5.1", "6.4.2", "6.5.3", "110.3.4", "1.1.5"}}, want: "110.3.4", wantErr: false},
		{name: "ValidMajorWithString", args: args{tags: []string{"v2.5.1", "v6.4.2", "v6.5.3", "v110.3.4", "v1.1.5"}}, want: "v110.3.4", wantErr: false},
		{name: "ValidReleaseCandidate", args: args{tags: []string{"v2.5.1-rc1", "v2.5.1-rc4", "v2.5.1-rc2", "v2.5.1-rc3", "v2.5.1-rc0"}}, want: "v2.5.1-rc4", wantErr: false},
		{name: "InvalidNoTags", args: args{tags: []string{}}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tagInput string
			if len(tt.args.tags) < 1 {
				tagInput = "*"
			} else {
				tagInput = tt.args.tags[len(tt.args.tags)-1]
			}
			regx, err := GetRegexExprForTag(tagInput)
			if err != nil {
				t.Errorf("GetRegexExprForTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := GetLatestTagWithRegexExpr(tt.args.tags, regx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestTagWithRegexExpr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetLatestTagWithRegexExpr() got = %v, want %v", got, tt.want)
			}
		})
	}
}
