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

package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var (
	UnknownAPIResponseError = errors.New("API error")
	StatusUnauthorizedError = errors.New("403")
	StatusForbiddenError    = errors.New("401")
	StatusToManyRequests    = errors.New("429")
)

const (
	httpAuthenticateHeader = "WWW-Authenticate"
	bearerRealmRegex       = "^Bearer realm=\"(.*?)\",service=\"(.*?)\"$"
	dockerRegistryVersion  = "v2"
)

type bearerToken struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}

type tagList struct {
	Tags []string `json:"tags"`
}

// OciImage is the interface that wraps a Image representation and its required information in a formatted format
// that the client requires for different kind of API calls
type OciImage interface {
	GetNameWithoutRegistry() string
	GetRegistryURL() string
}

// OciPullSecret is the interface that wraps a pull secret representation with a username and password.
type OciPullSecret interface {
	GetUsername() string
	GetPassword() string
}

// OciAPIClient requests a  registry of a given Image. If  pull secret is nil it will request the registry without basic-auth.
// The stores a bearer token to avoid unnecessary traffic and registry restrictions of max login. If an API call code is 401 or 403
// the client return a PermissionsError, else a ClientAPIError.
type OciAPIClient struct {
	Image       OciImage
	bearerToken string
	http.Client
}

// GetTagsForImage for configured client. If secret is nil the request will omit the BasicAuth HTTP header
func (c *OciAPIClient) GetTagsForImage(ctx context.Context, secret OciPullSecret) ([]string, error) {
	if c.bearerToken == "" {
		err := c.getBearerToken(ctx, secret)
		if err != nil {
			return nil, err
		}
		return c.getTags(ctx)
	}

	tags, err := c.getTags(ctx)
	if errors.Is(err, StatusUnauthorizedError) || errors.Is(err, StatusForbiddenError) {
		err := c.getBearerToken(ctx, secret)
		if err != nil {
			return nil, err
		}
		return c.getTags(ctx)
	}
	return tags, err
}

func (c *OciAPIClient) getBearerToken(ctx context.Context, secret OciPullSecret) error {
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

func (c *OciAPIClient) getRealmURLFromImageRegistry(ctx context.Context) (url string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+c.Image.GetRegistryURL()+"/"+dockerRegistryVersion+"/", nil)
	if err != nil {
		return "", fmt.Errorf("registries/api error: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return "", fmt.Errorf("registries/api error: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	if resp.StatusCode != http.StatusUnauthorized {
		return "", fmt.Errorf("registries/api %w: invalid response code %d from %s registries when trying to get realm URL for bearer token for Image %s. Registry does not follow the %s header standard. %d is required", UnknownAPIResponseError, resp.StatusCode, c.Image.GetRegistryURL(), c.Image.GetNameWithoutRegistry(), httpAuthenticateHeader, http.StatusUnauthorized)
	}

	respHeader := resp.Header.Get(httpAuthenticateHeader)
	if respHeader == "" {
		return "", fmt.Errorf("registry/api %w: Header \"%s\" is empty for requested url \"%s\"", UnknownAPIResponseError, httpAuthenticateHeader, c.Image.GetRegistryURL())
	}

	if !isValidHeader(bearerRealmRegex, respHeader) {
		return "", fmt.Errorf("\"registry/api %w: \"%s\" header does not contain any bearer realm information", UnknownAPIResponseError, httpAuthenticateHeader)
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

func (c *OciAPIClient) generateRealmURLWithService(realm, service string) string {
	return fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", realm, service, c.Image.GetNameWithoutRegistry())
}

func (c *OciAPIClient) getBearerTokenFromRealm(ctx context.Context, realmURL string, secret OciPullSecret) (token string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, realmURL, nil)
	if err != nil {
		return "", fmt.Errorf("registries/api %w: %s", UnknownAPIResponseError, err)
	}

	if reflect.ValueOf(secret).Kind() == reflect.Ptr && !reflect.ValueOf(secret).IsNil() {
		req.SetBasicAuth(secret.GetUsername(), secret.GetUsername())
	}

	resp, err := c.Do(req)
	if err != nil {
		return "", fmt.Errorf("registries/api %w: %s", UnknownAPIResponseError, err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	if err := handleResponseCodeOfResponse(resp); err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("registries/api %w: %s", UnknownAPIResponseError, err)
	}
	var t bearerToken
	if err := json.Unmarshal(body, &t); err != nil {
		return "", fmt.Errorf("registries/api %w: %s", UnknownAPIResponseError, err)
	}

	if t.Token == "" {
		return "", fmt.Errorf("registry/api %w: bearer token is empty in response %+v", UnknownAPIResponseError, t)
	}
	return t.Token, nil
}

func (c *OciAPIClient) getTags(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.generateGetTagsURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("registries/api %w: %s", UnknownAPIResponseError, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registries/api %w: %s", UnknownAPIResponseError, err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	if err := handleResponseCodeOfResponse(resp); err != nil {
		c.bearerToken = ""
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("registries/api %w: %s", UnknownAPIResponseError, err)
	}

	var tags tagList
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, fmt.Errorf("registries/api %w: %s", UnknownAPIResponseError, err)
	}

	return tags.Tags, nil
}

func (c *OciAPIClient) generateGetTagsURL() string {
	return fmt.Sprintf("https://%s/%s/%s/tags/list", c.Image.GetRegistryURL(), dockerRegistryVersion, c.Image.GetNameWithoutRegistry())
}

const (
	urlRegex     = "https://[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)"
	serviceRegex = "\"(.*?)\""
)

func handleResponseCodeOfResponse(resp *http.Response) error {
	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return fmt.Errorf("registries/api status %w on requesting %s, please check your permissions", StatusUnauthorizedError, resp.Request.URL.String())
	case resp.StatusCode == http.StatusForbidden:
		return fmt.Errorf("registries/api status %w on requesting %s, please check your permissions", StatusForbiddenError, resp.Request.URL.String())
	case resp.StatusCode == http.StatusTooManyRequests:
		return fmt.Errorf("registries/api status %w on requesting %s, please check your permissions", StatusToManyRequests, resp.Request.URL.String())
	case resp.StatusCode >= 300:
		return fmt.Errorf("registries/api %w: status %s on requesting %s", UnknownAPIResponseError, resp.Status, resp.Request.URL.String())
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
		return "", fmt.Errorf("registries/api %w: header '%s' does not contain a valid URL", UnknownAPIResponseError, header)
	}
	return url, nil
}

func extractService(header string) (string, error) {
	r := regexp.MustCompile(serviceRegex)
	service := r.FindString(header)
	if service == "" {
		return "", fmt.Errorf("registries/api %w: header '%s' does not contain a valid URL", UnknownAPIResponseError, header)
	}
	return strings.Replace(service, "\"", "", -1), nil
}
