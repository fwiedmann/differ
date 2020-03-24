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
	"strings"
	"time"

	"github.com/fwiedmann/differ/pkg/image"
)

const (
	httpAuthenticateHeader = "WWW-Authenticate"
	bearerRealmRegex       = "^Bearer realm=\"(.*?)\",service=\"(.*?)\"$"
	dockerRegistryVersion  = "v2"
)

type HTTPClient interface {
	MakeRequest(r *http.Request) (*http.Response, error)
}

type bearerToken struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}

type tagList struct {
	Tags []string `json:"tags"`
}

type Client struct {
	image       image.WithAssociatedPullSecrets
	bearerToken string
	http        HTTPClient
}

func New(image image.WithAssociatedPullSecrets, c HTTPClient) *Client {
	return &Client{
		image: image,
		http:  c,
	}
}

func (c *Client) GetTagsForImage(ctx context.Context, secret image.PullSecret) ([]string, error) {
	if c.bearerToken == "" {
		err := c.getBearerToken(ctx, secret)
		if err != nil {
			return nil, err
		}
	}
	return c.getTags(ctx)
}

func (c *Client) getBearerToken(ctx context.Context, secret image.PullSecret) error {
	realmURL, err := c.getRealmURLFromImageRegistry(ctx)
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

func (c *Client) getRealmURLFromImageRegistry(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+c.image.GetRegistryURL()+"/"+dockerRegistryVersion, nil)
	if err != nil {
		return "", newAPIErrorF(err, "registries/api error: %s", err)
	}

	resp, err := c.http.MakeRequest(req)
	if err != nil {
		return "", newAPIErrorF(err, "registries/api error: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		return "", newAPIErrorF(err, "registries/api error: invalid response code %d from %s registries when trying to get realm URL for bearer token for image %s. Registry does not follow the %s header standard. %d is required", resp.StatusCode, c.image.GetRegistryURL(), c.image.GetNameWithoutRegistry(), httpAuthenticateHeader, http.StatusUnauthorized)
	}

	respHeader := resp.Header.Get(httpAuthenticateHeader)
	if respHeader == "" {
		return "", newAPIErrorF(nil, "Header \"%s\" is empty for requested url \"%s\"", httpAuthenticateHeader, c.image.GetRegistryURL())
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

	return c.generateRealmURLWithService(realm, service), nil
}

func (c *Client) generateRealmURLWithService(realm, service string) string {
	return fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", realm, service, c.image.GetNameWithoutRegistry())
}

func (c *Client) getBearerTokenFromRealm(ctx context.Context, realmURL string, secret image.PullSecret) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, realmURL, nil)
	if err != nil {
		return "", newAPIErrorF(err, "registries/api error: %s", err)
	}

	if !secret.IsEmpty() {
		req.SetBasicAuth(secret.Username, secret.Username)
	}

	resp, err := c.http.MakeRequest(req)
	if err != nil {
		return "", newAPIErrorF(err, "registries/api error: %s", err)
	}
	defer resp.Body.Close()

	if err := handleResponseCodeOfResponse(resp); err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", newAPIErrorF(err, "registries/api error: %s", err)
	}
	var token bearerToken
	if err := json.Unmarshal(body, &token); err != nil {
		return "", newAPIErrorF(err, "registries/api error: %s", err)
	}
	return token.Token, nil
}

func (c *Client) getTags(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.generateGetTagsURL(), nil)
	if err != nil {
		return nil, newAPIErrorF(err, "registries/api error: %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)

	resp, err := c.http.MakeRequest(req)
	if err != nil {
		return nil, newAPIErrorF(err, "registries/api error: %s", err)
	}
	defer resp.Body.Close()

	if err := handleResponseCodeOfResponse(resp); err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newAPIErrorF(err, "registries/api error: %s", err)
	}

	var tags tagList
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, newAPIErrorF(err, "registries/api error: %s", err)
	}

	return tags.Tags, nil
}

func (c *Client) generateGetTagsURL() string {
	return fmt.Sprintf("https://%s/%s/%s/tags/list", c.image.GetRegistryURL(), dockerRegistryVersion, c.image.GetNameWithoutRegistry())
}
