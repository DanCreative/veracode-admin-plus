package veracode

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/sirupsen/logrus"
)

// GetAggregatedUsers returns a list of users with each of their roles
func (c *Client) GetAggregatedUsers(page int, size int, userType string) ([]*models.User, error) {
	summaryUsers, err := c.GetUsers(page, size, userType)
	if err != nil {
		return nil, err
	}

	userOrder := make(map[string]int, len(summaryUsers))
	aggregatedUsers := make([]*models.User, len(summaryUsers))

	for k, v := range summaryUsers {
		userOrder[v.UserId] = k
	}

	var wg sync.WaitGroup
	ch := make(chan *models.User)

	for _, user := range summaryUsers {
		wg.Add(1)
		go func(user models.User) {
			defer wg.Done()
			user, err := c.GetUser(user.UserId)
			if err != nil {
				return
			}
			ch <- &user
		}(*user)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for user := range ch {
		aggregatedUsers[userOrder[user.UserId]] = user
	}

	return aggregatedUsers, nil
}

// GetUsers fetches a summary of the users
func (c *Client) GetUsers(page int, size int, userType string) ([]*models.User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%susers/search?page=%d&size=%d&detailed=true&user_type=%s", c.BaseURL, page-1, size, userType), nil)
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

// GetUser fetches a user by ID
func (c *Client) GetUser(userId string) (models.User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%susers/%s?detailed=true", c.BaseURL, userId), nil)
	if err != nil {
		logrus.Error(err)
		return models.User{}, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		logrus.Error(err)
		return models.User{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		return models.User{}, err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("API error. http code: %v. Response Body: %s", resp.Status, string(body))
		logrus.Error(err)
		return models.User{}, err
	}

	var user models.User

	err = json.Unmarshal(body, &user)
	if err != nil {
		logrus.Error(err)
		return models.User{}, err
	}

	return user, nil
}
