package veracode

import (
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	BaseURL *url.URL
	Client  *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	baseEndpoint, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(baseEndpoint.Path, "/") {
		baseEndpoint.Path += "/"
	}

	return &Client{BaseURL: baseEndpoint, Client: httpClient}, nil
}
