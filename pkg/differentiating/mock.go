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

package differentiating

import "context"

// MockService implements the Service interface and should only be used for mocking
type MockService struct {
	AddErr, DeleteErr, UpdateErr, ListErr error
	ListResp                              []Image
}

// AddImage implements the Service interface
func (ms MockService) AddImage(_ context.Context, _ Image) error {
	return ms.AddErr
}

// DeleteImage implements the Service interface
func (ms MockService) DeleteImage(_ context.Context, _ Image) error {
	return ms.DeleteErr
}

// UpdateImage implements the Service interface
func (ms MockService) UpdateImage(_ context.Context, _ Image) error {
	return ms.UpdateErr
}

// ListImages implements the Service interface
func (ms MockService) ListImages(_ context.Context, _ ListOptions) ([]Image, error) {
	return ms.ListResp, ms.ListErr
}

// Notify implements the Service interface
func (ms MockService) Notify(event chan<- NotificationEvent) {
	go func() {
		for _, img := range ms.ListResp {
			event <- NotificationEvent{
				Image:  img,
				NewTag: "187",
			}
		}
	}()
}
