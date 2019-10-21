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
	httpClient "github.com/fwiedmann/differ/pkg/http"
	log "github.com/sirupsen/logrus"
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
	bearerToken  BearerToken
	RemoteLogger *log.Entry
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

type Remotes map[string]*Remote

func (r Remotes) CreateRemoteIfNotExists(image string) error {
	if _, found := r[image]; found {
		log.Debugf("Remote for image %s already exists", image)
	}
	remote, err := newRemote(image)
	if err != nil {
		return err
	}

	r[image] = remote
	return nil
}

// Error type for registry package
type Error struct {
	remoteURL string
	message   string
}

// Error to string
func (e Error) Error() string {
	return fmt.Sprintf("Remote %s error: %s,", e.remoteURL, e.message)
}

// NewError helper method to create a registry pkg error
func NewError(remoteURL, errorMessage string) Error {
	return Error{
		message:   errorMessage,
		remoteURL: remoteURL,
	}
}

// NewRemote inits a new remote
func newRemote(image string) (*Remote, error) {
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
		RemoteLogger: log.WithField("Remote", "Remote:"+parsedURL.String()),
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

	_, statusCode, header, err := httpClient.MakeRequest(http.MethodGet, basicRemoteURL)

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
	body, _, _, err := httpClient.MakeRequest(http.MethodGet, authRealmURL)
	if err != nil {
		return BearerToken{}, err
	}
	var token BearerToken
	if err := json.Unmarshal(body, &token); err != nil {
		return BearerToken{}, err
	}
	return token, nil
}

// GetTags get all available tags from remote
func (r *Remote) GetTags() ([]string, error) {
	// ToDo: check resp code, parse body, if bearer token is expired retry to get an new
	body, respCode, _, err := httpClient.MakeRequestWithHeader(http.MethodGet, r.URL.String()+"/tags/list", map[string]string{
		"Authorization": "Bearer " + r.bearerToken.Token,
	})
	if err != nil {
		return []string{}, err
	}

	// renew bearer token
	if respCode == http.StatusUnauthorized {
		r.bearerToken, _ = getBearerToken(r.authRealmURL)
		body, respCode, _, err = httpClient.MakeRequestWithHeader(http.MethodGet, r.URL.String()+"/tags/list", map[string]string{
			"Authorization": "Bearer " + r.bearerToken.Token,
		})
		if err != nil {
			return []string{}, err
		}
	}
	if respCode >= http.StatusBadRequest {
		return []string{}, NewError(r.URL.String(), fmt.Sprintf("Could not get tags from remote, statuscode: %d", respCode))
	}
	var list ListTagsResponse

	if err := json.Unmarshal(body, &list); err != nil {
		return []string{}, NewError(r.URL.String(), err.Error())
	}
	r.RemoteLogger.Tracef("Latest tags %v", list.Tags)
	return list.Tags, nil
}
