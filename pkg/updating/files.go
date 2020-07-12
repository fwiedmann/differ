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
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/fwiedmann/differ/pkg/analyzing"
)

func NewFileUpdatingService(dir string, getFilesFunc func() ([]io.ReadWriter, error)) (Service, error) {
	return &fileUpdatingService{
		contextDir: dir,
		getFiles:   getFilesFunc,
	}, nil
}

type fileUpdatingService struct {
	contextDir string
	getFiles   func() ([]io.ReadWriter, error)
}

func (g *fileUpdatingService) Update(_ context.Context, image Image) (Count, error) {
	expr, err := analyzing.GetRegexExprForTag(image.GetGetCompleteName())
	if err != nil {
		return 0, err
	}

	files, err := g.getFiles()
	if err != nil {
		return 0, err
	}

	return replaceImageInFilesByRegexExp(image, expr, files)
}

func replaceImageInFilesByRegexExp(image Image, regexExp *regexp.Regexp, files []io.ReadWriter) (Count, error) {
	var count Count
	for _, file := range files {
		lineCount, err := replaceImageInFileByRegexExp(image, regexExp, file)
		if err != nil {
			return 0, err
		}
		count = +lineCount
	}
	return count, nil
}

func replaceImageInFileByRegexExp(image Image, regexExp *regexp.Regexp, file io.ReadWriter) (Count, error) {
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(content), "\n")
	var count Count
	for i, line := range lines {
		updatedLine := regexExp.ReplaceAllString(line, image.GetGetCompleteName())
		if line != updatedLine {
			lines[i] = updatedLine
			count++
		}
	}

	_, err = fmt.Fprint(file, []byte(strings.Join(lines, "\n")))
	if err != nil {
		return count, err
	}
	return count, nil
}
