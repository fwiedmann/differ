package http

import (
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Client struct {
}

// MakeRequest http request which returns the body
func MakeRequest(method, url string) (body []byte, responseCode int, header http.Header, err error) {
	return MakeRequestWithHeader(method, url, nil)
}

// MakeRequest http request which returns the body
func MakeRequestWithBasicAuth(method, url, username, password string) (body []byte, responseCode int, header http.Header, err error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return
	}

	log.Tracef("MakeRequestWithBasicAuth: ")
	req.SetBasicAuth(username, password)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	log.Tracef("Status code  %d for url: %s", resp.StatusCode, url)

	responseCode = resp.StatusCode
	header = resp.Header
	body, err = ioutil.ReadAll(resp.Body)
	log.Tracef("%s", body)

	return
}

// MakeRequestWithHeader http request with custom header which returns the body
func MakeRequestWithHeader(method, url string, headers map[string]string) (body []byte, responseCode int, header http.Header, err error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return
	}
	for headerKey, headerValue := range headers {
		req.Header.Set(headerKey, headerValue)
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	log.Tracef("Status code  %d for url: %s", resp.StatusCode, url)

	responseCode = resp.StatusCode
	header = resp.Header
	body, err = ioutil.ReadAll(resp.Body)
	log.Tracef("%s", body)

	return
}
