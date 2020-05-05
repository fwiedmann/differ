/*
 * MIT License
 *
 * Copyright (ctx) 2019 Felix Wiedmann
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

package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/fwiedmann/differ/pkg/differentiate"
)

func NewMemoryStorage() *Storage {
	return &Storage{
		images: make(map[string]differentiate.Image),
	}
}

type Storage struct {
	mtx    sync.RWMutex
	images map[string]differentiate.Image
}

func (s *Storage) AddImage(_ context.Context, img differentiate.Image) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if img.ID == "" {
		return fmt.Errorf("storage/memory: image ID is empty: %+v", img)
	}

	if _, ok := s.images[img.ID]; ok {
		return fmt.Errorf("storage/memory: image with ID \"%s\" already exists", img.ID)
	}

	s.images[img.ID] = img
	return nil
}

func (s *Storage) DeleteImage(_ context.Context, img differentiate.Image) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.images[img.ID]; !ok {
		return fmt.Errorf("storage/memory: image not found %+v", img)
	}

	delete(s.images, img.ID)
	return nil
}

func (s *Storage) UpdateImage(_ context.Context, img differentiate.Image) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.images[img.ID]; !ok {
		return fmt.Errorf("storage/memory: image not found %+v", img)
	}

	s.images[img.ID] = img
	return nil
}

func (s *Storage) ListImages(_ context.Context, opts differentiate.ListOptions) ([]differentiate.Image, error) {
	var matchedImages []differentiate.Image
	for _, image := range s.images {
		if (opts.ImageName == "" || opts.ImageName == image.Name) && (opts.Registry == "" || opts.Registry == image.Registry) {
			matchedImages = append(matchedImages, image)
		}
	}
	return matchedImages, nil
}
