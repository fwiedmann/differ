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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/fwiedmann/differ/pkg/image"
)

const (
	httpAuthenticateHeader = "WWW-Authenticate"
	bearerRealmRegex       = "^Bearer realm=\"(.*?)\",service=\"(.*?)\"$"
	urlRegex               = "https://[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)"
	serviceRegex           = "\"(.*?)\""
	dockerRegistryVersion  = "v2"
)

type httpClient interface {
	MakeRequest(r *http.Request) (*http.Response, error)
}

type BearerToken struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}

type Client struct {
	image       string
	registryURL string
	bearerToken string
	http        httpClient
}

func New(image, registryURL string) *Client {
	return &Client{
		image:       image,
		registryURL: registryURL,
	}
}

func (c *Client) GetTagsForImage(ctx context.Context, secret image.PullSecret) ([]string, error) {
	if c.bearerToken == "" {
		err := c.getBearerToken(ctx, secret)
		if err != nil {
			return nil, err
		}
	}
	// ToDo: implement get Tags
	return c.getTag()
}

func (c *Client) getBearerToken(ctx context.Context, secret image.PullSecret) error {
	realmURL, err := c.getRealmURLFromHeader(ctx)
	if err != nil {
		return err
	}

	token, err := c.getBearerTokenFromRealm(ctx, realmURL, secret)
	if err != nil {
		return err
	}

	c.bearerToken = token
	return nil
}

func (c *Client) getRealmURLFromHeader(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.registryURL+"/"+dockerRegistryVersion, nil)
	if err != nil {
		return "", newAPIErrorF(err, "registry/api error")
	}

	resp, err := c.http.MakeRequest(req)
	if err != nil {
		return "", newAPIErrorF(err, "registry/api error")
	}

	if resp.StatusCode != http.StatusUnauthorized {
		return "", newAPIErrorF(err, "registry/api error: invalid response code %d from %s registry when trying to get realm URL for bearer token for image %s. Registry does not follow the %s header standard. %d is required", resp.StatusCode, c.registryURL, c.image, httpAuthenticateHeader, http.StatusUnauthorized)
	}

	respHeader := resp.Header.Get(httpAuthenticateHeader)
	if respHeader == "" {
		return "", newAPIErrorF(nil, "Header \"%s\" is empty for requested url \"%s\"", httpAuthenticateHeader, c.registryURL)
	}

	if !isValidHeader(bearerRealmRegex, respHeader) {
		return "", newAPIErrorF(nil, "\"%s\" header does not contain any bearer realm information", httpAuthenticateHeader)
	}

	headerValues := strings.Split(respHeader, ",")
	realm, err := extractRealmURL(headerValues[0])
	if err != nil {
		return "", err
	}

	service, err := extractService(headerValues[1])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", realm, service, c.image), nil
}

func (c *Client) getBearerTokenFromRealm(ctx context.Context, realmURL string, secret image.PullSecret) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, realmURL, nil)

	if err != nil {
		return "", newAPIErrorF(err, "registry/api error")
	}

	if !secret.IsEmpty() {
		req.SetBasicAuth(secret.Username, secret.Username)
	}

	resp, err := c.http.MakeRequest(req)
	if err != nil {
		return "", newAPIErrorF(err, "registry/api error")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", newAPIErrorF(err, "registry/api error")
	}
	var token BearerToken
	if err := json.Unmarshal(body, &token); err != nil {
		return "", err
	}
	return token.Token, nil
}

func (c *Client) getTag() ([]string, error) {
	return nil, nil
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
