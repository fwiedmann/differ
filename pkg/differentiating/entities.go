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

import "fmt"

type PullSecret struct {
	Username string
	Password string
}

func (p *PullSecret) GetUsername() string {
	return p.Username
}

func (p *PullSecret) GetPassword() string {
	return p.Password
}

type Image struct {
	ID       string
	Registry string
	Name     string
	Tag      string
	Auth     []*PullSecret
}

func (i Image) GetNameWithoutRegistry() string {
	return i.Name
}

func (i Image) GetNameWithRegistry() string {
	return fmt.Sprintf("%s/%s", i.Registry, i.Name)
}

func (i Image) GetRegistryURL() string {
	return i.Registry
}

type ListOptions struct {
	ImageName string
	Registry  string
}

type NotificationEvent struct {
	Image  Image
	NewTag string
}
