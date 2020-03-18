package http

import (
	"net/http"
	http2 "net/http"
)

type Client struct {
}

func (c Client) MakeRequest(r *http.Request) (*http.Response, error) {
	client := http2.Client{}
	return client.Do(r)
}
