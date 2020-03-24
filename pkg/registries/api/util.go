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
	"net/http"
	"regexp"
	"strings"
)

const (
	urlRegex     = "https://[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)"
	serviceRegex = "\"(.*?)\""
)

func handleResponseCodeOfResponse(resp *http.Response) error {
	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return newPermissionsError(nil, "registries/api status %s on requesting %s, please check your permissions", resp.Status, resp.Request.URL.String())
	case resp.StatusCode == http.StatusForbidden:
		return newPermissionsError(nil, "registries/api status %s on requesting %s, please check your permissions", resp.Status, resp.Request.URL.String())
	case resp.StatusCode >= 300:
		return newAPIErrorF(nil, "registries/api status %s on requesting %s", resp.Status, resp.Request.URL.String())
	default:
		return nil
	}
}

func isValidHeader(headerRegex, header string) bool {
	r := regexp.MustCompile(headerRegex)
	return r.MatchString(header)
}

func extractRealmURL(header string) (string, error) {
	r := regexp.MustCompile(urlRegex)
	url := r.FindString(header)
	if url == "" {
		return "", newAPIErrorF(nil, "header '%s' does not contain a valid URL", header)
	}
	return url, nil
}

func extractService(header string) (string, error) {
	r := regexp.MustCompile(serviceRegex)
	service := r.FindString(header)
	if service == "" {
		return "", newAPIErrorF(nil, "header '%s' does not contain a valid URL", header)
	}
	return strings.Replace(service, "\"", "", -1), nil
}
