package http

import (
	"io/ioutil"
	"net/http"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

// MakeRequest http request which returns the body
func MakeRequest(method, url string) (body []byte, responseCode int, header http.Header, err error) {
	return MakeRequestWithHeader(method, url, nil)
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
	defer resp.Body.Close()
	if err != nil {
		return
	}

	responseCode = resp.StatusCode
	header = resp.Header
	body, err = ioutil.ReadAll(resp.Body)

	return
}
