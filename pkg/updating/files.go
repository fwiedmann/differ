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
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/fwiedmann/differ/pkg/analyzing"
)

func NewFileUpdatingService(dir string, isFileValidToCheckForUpdatesFun func(name string) bool, logger logrus.Logger) (Service, error) {

	return &fileUpdatingService{
		contextDir:                      dir,
		isFileValidToCheckForUpdatesFun: isFileValidToCheckForUpdatesFun,
		log:                             logger,
	}, nil
}

type fileUpdatingService struct {
	contextDir                      string
	isFileValidToCheckForUpdatesFun func(name string) bool
	log                             logrus.Logger
}

func (g *fileUpdatingService) Update(_ context.Context, image Image) (Count, error) {

	expr, err := analyzing.GetRegexExprForTag(image.GetGetCompleteName())
	if err != nil {
		return 0, err
	}

	files, err := g.getAllValidFilesForUpdating()
	if err != nil {
		return 0, err
	}

	return replaceImageInFilesByRegexExp(image, expr, files)
}

func (g *fileUpdatingService) getAllValidFilesForUpdating() ([]string, error) {
	var files []string
	err := filepath.Walk(g.contextDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if !g.isFileValidToCheckForUpdatesFun(path) {
			return nil
		}

		files = append(files, path)
		return nil
	})
	return files, err
}

func replaceImageInFilesByRegexExp(image Image, regexExp *regexp.Regexp, files []string) (Count, error) {
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

func replaceImageInFileByRegexExp(image Image, regexExp *regexp.Regexp, file string) (Count, error) {
	content, err := ioutil.ReadFile(file)
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
	return count, ioutil.WriteFile(file, []byte(strings.Join(lines, "\n")), 0600)
}
