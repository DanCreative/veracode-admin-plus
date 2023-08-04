package veracode

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/sirupsen/logrus"
)

// Hardcoded to page 1 size 40 for simplicity for now
// TODO: Automatic paging
func (c *Client) GetTeamsAsync(result chan any) {
	defer close(result)
	req, err := http.NewRequest("GET", fmt.Sprintf("%steams?page=0&size=40", c.BaseURL), nil)
	if err != nil {
		logrus.Error(err)
		result <- err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		logrus.Error(err)
		result <- err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		result <- err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("API error. http code: %v. Response Body: %s", resp.Status, string(body))
		logrus.Error(err)
		result <- err
	}

	data := struct {
		Embedded struct {
			Teams []models.Team `json:"teams"`
		} `json:"_embedded"`
	}{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		logrus.Error(err)
		result <- err
	}

	result <- data.Embedded.Teams
}
