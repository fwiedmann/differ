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
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fwiedmann/differ/pkg/metrics"

	"github.com/fwiedmann/differ/pkg/store"

	httpClient "github.com/fwiedmann/differ/pkg/http"
	log "github.com/sirupsen/logrus"
)

const (
	apiVersion    string = "v2"
	authHeaderKey string = "Www-Authenticate"
	dockerHubURL  string = "https://index.docker.io/"
)

type Remote struct {
	URL          *url.URL
	authRealmURL string
	bearerToken  BearerToken
	RemoteLogger *log.Entry
	Image        string
	auths        []store.ImagePullSecret
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

type Remotes struct {
	data map[string]*Remote
	m    sync.RWMutex
}

func NewRemoteStore() *Remotes {
	return &Remotes{
		data: make(map[string]*Remote),
		m:    sync.RWMutex{},
	}
}

func (r *Remotes) CreateOrUpdateRemote(image string, gatheredAuths []store.ImagePullSecret) error {
	r.m.Lock()
	defer r.m.Unlock()

	if remote, found := r.data[image]; found {
		log.Debugf("Remote for image %s already exists, only update auths", image)
		remote.auths = gatheredAuths
		return nil
	}
	remote, err := newRemote(image, gatheredAuths)
	if err != nil {
		return err
	}
	metrics.DifferRegistryTagError.WithLabelValues(image).Set(0)
	r.data[image] = remote
	return nil
}

func (r *Remotes) GetRemoteByID(image string) *Remote {
	r.m.RLock()
	defer r.m.RUnlock()

	return r.data[image]
}

// Error types for registry package
type Error struct {
	remoteURL string
	message   string
}

// Error to string
func (e Error) Error() string {
	return fmt.Sprintf("Remote %s error: %s", e.remoteURL, e.message)
}

// NewError helper method to create a registry pkg error
func NewError(remoteURL, errorMessage string) Error {
	return Error{
		message:   errorMessage,
		remoteURL: remoteURL,
	}
}

// NewRemote inits a new remote
func newRemote(image string, gatheredAuths []store.ImagePullSecret) (*Remote, error) {
	parsedURL, err := parseImageToURL(modifyIfDockerHubImage(image))
	if err != nil {
		return nil, NewError(image, fmt.Sprintf("Could no parse image to remote URL: %s", err.Error()))
	}

	realm, err := getAuthRealmURL(parsedURL)
	if err != nil {
		return nil, NewError(parsedURL.String(), "Could not get registry authURL. Error: "+err.Error())
	}

	return &Remote{
		URL:          parsedURL,
		authRealmURL: realm,
		RemoteLogger: log.WithField("Remote", "Remote:"+parsedURL.String()),
		Image:        image,
		auths:        gatheredAuths,
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

func getBearerToken(authRealmURL string, authSecret store.ImagePullSecret) (BearerToken, error) {
	var body []byte
	var err error
	var respCode int

	if authSecret.Username != "" && authSecret.Password != "" {
		body, respCode, _, err = httpClient.MakeRequestWithBasicAuth(http.MethodGet, authRealmURL, authSecret.Username, authSecret.Password)
		if err != nil {
			return BearerToken{}, err
		}
	} else {
		body, respCode, _, err = httpClient.MakeRequest(http.MethodGet, authRealmURL)
		if err != nil {
			return BearerToken{}, err
		}
	}

	if respCode >= http.StatusMultipleChoices {
		return BearerToken{}, NewError(authRealmURL, fmt.Sprintf("Could not get bearer token: status code %d", respCode))
	}

	var token BearerToken
	if err := json.Unmarshal(body, &token); err != nil {
		return BearerToken{}, err
	}
	return token, nil
}

func modifyIfDockerHubImage(image string) string {
	if !strings.Contains(image, ".") {
		return fmt.Sprintf("%s%s", dockerHubURL, image)
	}
	return image
}
func listTags(remoteURL, authToken string) ([]byte, int, error) {
	body, respCode, _, err := httpClient.MakeRequestWithHeader(http.MethodGet, remoteURL+"/tags/list", map[string]string{
		"Authorization": "Bearer " + authToken,
	})
	if err != nil {
		return []byte{}, 0, err
	}
	return body, respCode, nil
}

// GetTags get all available tags from remote
func (r *Remote) GetTags() ([]string, error) {
	var respBody []byte
	var list ListTagsResponse
	if r.bearerToken.Token == "" {
		// add empty auth for code reductions
		if len(r.auths) == 0 {
			r.auths = append(r.auths, store.ImagePullSecret{
				Username: "",
				Password: "",
			})
		}
		// trying to get tags from auhts, will break if successfully
		for _, auth := range r.auths {
			token, err := getBearerToken(r.authRealmURL, auth)
			if err != nil {
				r.RemoteLogger.Warnf("%s", err.Error())
				continue
			}
			body, respCode, err := listTags(r.URL.String(), token.Token)
			if err != nil {
				return []string{}, err
			}

			if respCode >= http.StatusMultipleChoices {
				if respCode == http.StatusUnauthorized {
					if strings.Contains(r.URL.String(), dockerHubURL) && !strings.Contains(r.URL.Path, "library") {
						r.RemoteLogger.Debugf("Trying to check if image is available under /library on docker hub")
						tmpRemote, err := newRemote("library/"+r.Image, r.auths)
						if err != nil {
							return []string{}, err
						}
						return tmpRemote.GetTags()
					}
				}
				continue
			}

			respBody = body
			// set token as default token
			r.bearerToken = token

			if err := json.Unmarshal(respBody, &list); err != nil {
				return []string{}, NewError(r.URL.String(), err.Error())
			}
			r.RemoteLogger.Tracef("Latest tags %v", list.Tags)
			return list.Tags, nil
		}
	} else {
		respBody, respCode, err := listTags(r.URL.String(), r.bearerToken.Token)
		if err != nil {
			return []string{}, err
		}

		// If provided token is not valid anymore, reset and call GetTags again.
		if respCode >= http.StatusMultipleChoices {
			r.RemoteLogger.Tracef("Reset Bearer Token")
			r.RemoteLogger.Tracef("Could not get tags : status code: %d", respCode)
			r.bearerToken = BearerToken{}

			return r.GetTags()
		}

		if err := json.Unmarshal(respBody, &list); err != nil {
			return []string{}, NewError(r.URL.String(), err.Error())
		}
		r.RemoteLogger.Tracef("Latest tags %v", list.Tags)
		return list.Tags, nil
	}
	metrics.DifferRegistryTagError.WithLabelValues(r.Image).Set(1)
	return []string{}, NewError(r.URL.String(), "Could not get tags")

}
