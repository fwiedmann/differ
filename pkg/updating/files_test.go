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

package updating

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestNewFileUpdatingService(t *testing.T) {
	type args struct {
		getFilesFun func() ([]string, error)
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{func() ([]string, error) {
				return nil, nil
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFileUpdatingService(tt.args.getFilesFun)

			fs, ok := got.(*fileUpdatingService)
			if !ok {
				t.Errorf("NewFileUpdatingService() = %v, want fileUpdatingService", fs)
			}
		})
	}
}

func Test_fileUpdatingService_Update(t *testing.T) {
	type prepare struct {
		imageOccurrenceInFle int
		getFilesFunErr       error
		imageInFile          Image
		skipTestFileCreation bool
	}

	type args struct {
		ctx   context.Context
		image Image
	}

	tests := []struct {
		name    string
		args    args
		prepare prepare
		want    Count
		wantErr bool
	}{
		{
			name: "UpdateWithCorrectCount",
			args: args{
				ctx: context.Background(),
				image: Image{
					Name: "wiedmannfelix/differ",
					Tag:  "1.9.2",
				},
			},
			prepare: prepare{
				imageOccurrenceInFle: 5,
				getFilesFunErr:       nil,
				imageInFile: Image{
					Name: "wiedmannfelix/differ",
					Tag:  "1.8.7",
				},
			},
			want:    5,
			wantErr: false,
		},
		{
			name: "GetFilesFunError",
			args: args{
				ctx: context.Background(),
				image: Image{
					Name: "wiedmannfelix/differ",
					Tag:  "1.9.2",
				},
			},
			prepare: prepare{
				imageOccurrenceInFle: 5,
				getFilesFunErr:       errors.New("getFilesFun() error"),
				imageInFile: Image{
					Name: "wiedmannfelix/differ",
					Tag:  "1.8.7",
				},
			},
			want:    5,
			wantErr: true,
		},
		{
			name: "FileNotFound",
			args: args{
				ctx: context.Background(),
				image: Image{
					Name: "wiedmannfelix/differ",
					Tag:  "1.9.2",
				},
			},
			prepare: prepare{
				skipTestFileCreation: true,
			},
			want:    5,
			wantErr: true,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := ""
			if !tt.prepare.skipTestFileCreation {
				var err error
				testFile, err = createTestFile(tt.prepare.imageInFile, tt.prepare.imageOccurrenceInFle)
				if err != nil {
					t.Errorf("createTestFile() err = %s", err)
					return
				}
				defer os.Remove(testFile)
			}

			fs := fileStorrer{
				fileName: testFile,
				err:      tt.prepare.getFilesFunErr,
			}

			g := &fileUpdatingService{
				getFiles: fs.getFiles,
			}

			got, err := g.Update(tt.args.ctx, tt.args.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if got != tt.want {
				t.Errorf("Update() got = %v, want %v", got, tt.want)
				return
			}

			if err := checkImageOccurrenceCountInFile(tt.args.image, testFile, tt.prepare.imageOccurrenceInFle); err != nil {
				t.Error(err)
				return
			}
		})
	}
}

func createTestFile(i Image, repeatCount int) (string, error) {
	f, err := ioutil.TempFile("", "testFile")
	if err != nil {
		return "", err
	}

	var content string
	for e := 0; e < repeatCount; e++ {
		content += fmt.Sprintf("\nTEST TEST TEST\n"+
			"%s:%s\n"+
			"TEST TEST TEST\n"+
			"%s:WRONGTAG",
			i.Name, i.Tag, i.Name)
	}

	if err := ioutil.WriteFile(f.Name(), []byte(content), 0777); err != nil {
		return "", err
	}

	return f.Name(), nil
}

func checkImageOccurrenceCountInFile(image Image, file string, expectedCount int) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	expr, err := regexp.Compile(image.GetGetCompleteName())
	if err != nil {
		return err
	}

	var actualCount int
	for _, line := range strings.Split(string(content), "\n") {
		for _, word := range strings.Split(line, " ") {
			if expr.MatchString(word) {
				actualCount += 1
			}
		}
	}

	if expectedCount != actualCount {
		return fmt.Errorf("image occurence want %d, got %d", expectedCount, actualCount)
	}
	return nil
}

type fileStorrer struct {
	fileName string
	err      error
}

func (f *fileStorrer) getFiles() ([]string, error) {
	return []string{f.fileName}, f.err
}
