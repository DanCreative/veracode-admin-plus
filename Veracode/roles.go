package veracode

import (
	"encoding/json"
	"os"

	"github.com/DanCreative/veracode-admin-plus/models"
)

var AddScanTypesRoles = map[string]bool{
	"extcreator":   true,
	"extseclead":   true,
	"extsubmitter": true,
}

func (c *Client) GetRoles() ([]models.Role, error) {
	var roles []models.Role
	// TODO: Update to make API call
	out, err := os.ReadFile(os.Getenv("API_RESPONSE_ROLES"))
	if err != nil {
		return nil, err
	}
	data := struct {
		Embedded struct {
			Roles []models.Role `json:"roles"`
		} `json:"_embedded"`
	}{}

	err = json.Unmarshal(out, &data)
	if err != nil {
		return nil, err
	}

	for _, role := range data.Embedded.Roles {
		if AddScanTypesRoles[role.RoleName] {
			role.IsAddScanTypes = true
		}
		if !role.IsApi {
			roles = append(roles, role)
		}
	}
	return roles, nil
}
