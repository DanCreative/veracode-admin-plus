package veracode

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/sirupsen/logrus"
)

var AddScanTypesRoles = map[string]bool{
	"extcreator":   true,
	"extseclead":   true,
	"extsubmitter": true,
}

// ! Hardcoded to page 1 size 40 for simplicity for now
// ! Only handling Human users at the moment
func (c *Client) GetRoles() error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%sroles?page=0&size=40", c.BaseURL), nil)
	if err != nil {
		logrus.Error(err)
		return err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		logrus.Error(err)
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		return err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("API error. http code: %v. Response Body: %s", resp.Status, string(body))
		logrus.Error(err)
		return err
	}

	data := struct {
		Embedded struct {
			Roles []models.Role `json:"roles"`
		} `json:"_embedded"`
	}{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		logrus.Error(err)
		return err
	}

	var humanRoles []models.Role

	// Filter out all of the API Roles
	for _, role := range data.Embedded.Roles {
		if AddScanTypesRoles[role.RoleName] {
			role.IsAddScanTypes = true
		}
		if !role.IsApi {
			humanRoles = append(humanRoles, role)
		}
	}

	c.Roles = humanRoles
	return nil
}
