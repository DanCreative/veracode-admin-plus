package veracode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/sirupsen/logrus"
)

// GetAggregatedUsers returns a list of users with each of their roles
func (c *Client) GetAggregatedUsers(page int, size int, userType string) ([]*models.User, models.PageMeta, error) {
	summaryUsers, meta, err := c.GetUsers(page, size, userType)
	if err != nil {
		return nil, models.PageMeta{}, err
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
		logrus.Debugf("Successfully got user(%s): %v", user.UserId, user)
		aggregatedUsers[userOrder[user.UserId]] = user
	}

	return aggregatedUsers, meta, nil
}

// GetUsers fetches a summary of the users
func (c *Client) GetUsers(page int, size int, userType string) ([]*models.User, models.PageMeta, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%susers/search?page=%d&size=%d&detailed=true&user_type=%s", c.BaseURL, page-1, size, userType), nil)
	if err != nil {
		logrus.Error(err)
		return nil, models.PageMeta{}, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		logrus.Error(err)
		return nil, models.PageMeta{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		return nil, models.PageMeta{}, err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("API error. http code: %v. Response Body: %s", resp.Status, string(body))
		logrus.Error(err)
		return nil, models.PageMeta{}, err
	}

	userSummaries := struct {
		Embedded struct {
			Users []*models.User `json:"users"`
		} `json:"_embedded"`
		Page models.PageMeta `json:"page"`
	}{}

	err = json.Unmarshal(body, &userSummaries)
	if err != nil {
		logrus.Error(err)
		return nil, models.PageMeta{}, err
	}

	return userSummaries.Embedded.Users, userSummaries.Page, nil
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

// PutPartialUser updates a user by ID
func (c *Client) PutPartialUser(userId string, user models.User) error {
	reqBody, err := json.Marshal(user)
	if err != nil {
		logrus.Error(err)
		return err
	}
	logrus.Debug(string(reqBody))

	req, err := http.NewRequest("PUT", fmt.Sprintf("%susers/%s?partial=true", c.BaseURL, userId), bytes.NewBuffer(reqBody))
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

	logrus.Infof("Successfully updated user: %s", userId)
	return nil
}

// BulkPutPartialUsers updates multiple users async
func (c *Client) BulkPutPartialUsers(users map[string]models.User) []error {
	chError := make(chan error, len(users))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for k, v := range users {
		wg.Add(1)
		go func(userId string, user models.User, ch chan error) {
			ch <- c.PutPartialUser(userId, user)
			wg.Done()

		}(k, v, chError)
	}
	go func() {
		wg.Wait()
		close(chError)
	}()

	var errors []error

	for err := range chError {
		if err != nil {
			logrus.Errorf("Error picked up during bulk put: %s", err)
			mu.Lock()
			errors = append(errors, err)
			mu.Unlock()
		}
	}
	return errors
}
