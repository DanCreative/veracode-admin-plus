package veracode

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/sirupsen/logrus"
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

func (c *Client) GetAggregatedUsers(page int, size int) ([]*models.User, error) {
	return c.GetUsers(page, size)
}

func (c *Client) GetUsers(page int, size int) ([]*models.User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%susers", c.BaseURL), nil)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("API error. http code: %v. Response Body: %s", resp.Status, string(body))
		logrus.Error(err)
		return nil, err
	}

	userSummaries := struct {
		Embedded struct {
			Users []*models.User `json:"users"`
		} `json:"_embedded"`
	}{}

	err = json.Unmarshal(body, &userSummaries)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	return userSummaries.Embedded.Users, nil
}
