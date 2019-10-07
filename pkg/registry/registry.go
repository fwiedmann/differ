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
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	apiVersion    string = "v2"
	authHeaderKey string = "Www-Authenticate"
)

type Remote struct {
	URL          *url.URL
	authRealmURL string
	client       http.Client
	bearerToken  BearerToken
}

type BearerToken struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}

type ListTagsResponse struct {
	Tags []string `json:"tags"`
}

// Error type for registry package
type Error struct {
	remoteURL string
	message   string
}

// Error to string
func (e Error) Error() string {
	return fmt.Sprintf("registy remote %s error: %s,", e.remoteURL, e.message)
}

// NewError helper method to create a registry pkg error
func NewError(remoteURL, errorMessage string) Error {
	return Error{
		message:   errorMessage,
		remoteURL: remoteURL,
	}
}

// NewRemote inits a new remote
func NewRemote(image string) (*Remote, error) {
	parsedURL, err := parseImageToURL(image)
	if err != nil {
		return nil, NewError(image, fmt.Sprintf("Could no parse image to remote URL: %s", err.Error()))
	}

	realm, err := getAuthRealmURL(parsedURL)
	if err != nil {
		return nil, NewError(parsedURL.String(), "Could not get registry authURL. Error: "+err.Error())
	}

	token, err := getBearerToken(realm)
	if err != nil {
		return nil, NewError(parsedURL.String(), "Could not bearer token. Error: "+err.Error())
	}

	return &Remote{
		URL:          parsedURL,
		authRealmURL: realm,
		bearerToken:  token,
	}, nil
}

func parseImageToURL(image string) (*url.URL, error) {

	if !strings.Contains(image, "https://") {
		image = fmt.Sprintf("https://%s", image)
	}

	parsedURL, err := url.Parse(image)
	if err != nil {
		return nil, err
	}

	parsedURL.Path = fmt.Sprintf("/%s%s", apiVersion, parsedURL.Path)

	return parsedURL, nil
}

func getAuthRealmURL(remoteURL *url.URL) (string, error) {
	basicRemoteURL := fmt.Sprintf("%s://%s/%s", remoteURL.Scheme, remoteURL.Hostname(), apiVersion)

	_, statusCode, header, err := makeRequest(http.MethodGet, basicRemoteURL)

	if err != nil {
		return "", err
	}
	if statusCode == http.StatusNotFound {
		return "", NewError(remoteURL.String(), "Could not resolve remote auth endpoint, will skip")
	}

	authHeaderValues := strings.Split(header[authHeaderKey][0], " ")
	if authHeaderValues[0] != "Bearer" {
		return "", NewError(remoteURL.String(), "Remotes auth does not support bearer auth, will skip.")
	}

	var realmURL string
	var service string

	bearerValues := strings.Split(authHeaderValues[1], ",")
	for _, value := range bearerValues {
		if strings.Contains(value, "realm") {
			realmURL = strings.ReplaceAll(strings.TrimLeft(value, "realm=\""), "\"", "")
		}
		if strings.Contains(value, "service") {
			service = strings.ReplaceAll(value, "\"", "")
		}
	}

	realmURL += fmt.Sprintf("?%s&scope=repository:%s:pull", service, strings.TrimLeft(remoteURL.Path, "v2/"))
	return realmURL, nil
}

func getBearerToken(authRealmURL string) (BearerToken, error) {
	body, _, _, err := makeRequest(http.MethodGet, authRealmURL)
	if err != nil {
		return BearerToken{}, err
	}
	var token BearerToken
	if err := json.Unmarshal(body, &token); err != nil {
		return BearerToken{}, err
	}
	return token, nil
}

// makeRequest helper method for http requests
func makeRequest(method, url string) (body []byte, responseCode int, header http.Header, err error) {
	return makeRequestWithHeader(method, url, nil)
}

func makeRequestWithHeader(method, url string, headers map[string]string) (body []byte, responseCode int, header http.Header, err error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return
	}
	for headerKey, headerValue := range headers {
		req.Header.Set(headerKey, headerValue)
	}
	client := http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return
	}

	responseCode = resp.StatusCode
	header = resp.Header
	body, err = ioutil.ReadAll(resp.Body)

	return
}

// GetTags get all available tags from remote
func (r *Remote) GetTags() ([]string, error) {
	// ToDo: check resp code, parse body, if bearer token is expired retry to get an new
	body, _, _, err := makeRequestWithHeader(http.MethodGet, r.URL.String()+"/tags/list", map[string]string{
		"Authorization": "Bearer " + r.bearerToken.Token,
	})
	if err != nil {
		return []string{}, NewError(r.URL.String(), err.Error())
	}

	var list ListTagsResponse

	if err := json.Unmarshal(body, &list); err != nil {
		return []string{}, NewError(r.URL.String(), err.Error())
	}

	log.Debugf("%s Tags: %+v", r.URL, list)
	return list.Tags, nil
}
