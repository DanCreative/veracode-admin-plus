package veracode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/sirupsen/logrus"
)

// GetAggregatedUsers returns a list of users with each of their roles
func (c *Client) GetAggregatedUsers(urlArgs url.Values) ([]*models.User, models.PageMeta, error) {
	summaryUsers, meta, err := c.GetUsers(urlArgs)
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
		aggregatedUsers[userOrder[user.UserId]] = user
	}

	return aggregatedUsers, meta, nil
}

// GetUsers fetches a summary of the users
func (c *Client) GetUsers(urlArgs url.Values) ([]*models.User, models.PageMeta, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%susers/search", c.BaseURL), nil)
	if err != nil {
		logrus.Error(err)
		return nil, models.PageMeta{}, err
	}

	req.URL.RawQuery = urlArgs.Encode()

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
	url := fmt.Sprintf("%susers/%s?partial=true", c.BaseURL, userId)
	method := "PUT"

	reqBody, err := json.Marshal(user)
	if err != nil {
		logrus.Error(err)
		return &UserError{err: err}
	}
	//logrus.Debug(string(reqBody))

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		logrus.Error(err)
		return &UserError{err: err}
	}

	logrus.WithFields(logrus.Fields{"Function": "PutPartialUser"}).Infof("%s %s", req.Method, req.URL)

	resp, err := c.Client.Do(req)
	if err != nil {
		logrus.Error(err)
		return &UserError{err: err}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		return &UserError{err: err}
	}

	if resp.StatusCode != 200 {
		uerr := UserError{}
		err := json.Unmarshal(body, &uerr)
		if err != nil {
			logrus.Error(err)
			return &UserError{err: err}
		}

		uerr.Method = method
		uerr.Url = url
		uerr.UserId = userId

		logrus.Error(uerr)
		return &uerr
	}

	logrus.Infof("Successfully updated user: %s", userId)
	return nil
}

// BulkPutPartialUsers updates multiple users async
func (c *Client) BulkPutPartialUsers(users map[string]models.User) []error {
	logrus.WithFields(logrus.Fields{"Function": "BulkPutPartialUsers"}).Trace("Start")
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
