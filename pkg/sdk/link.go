package sdk

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/google/uuid"
)

func (s *AccountService) linkHelper(method, username, id string, config interface{}) error {
	baseURL, err := url.Parse(s.Client.API + "/account/" + username)
	if err != nil {
		return err
	}

	linkRequest := &struct {
		Config interface{} `json:"config"`
		Value  string      `json:"value"`
	}{
		Config: config,
		Value:  id,
	}
	body, err := json.Marshal(linkRequest)
	if err != nil {
		return err
	}

	_, err = s.Client.request(method, baseURL.String(), body)
	if err != nil {
		return err
	}
	return nil
}

func (s *AccountService) AddWebhook() (string, error) {
	pin := uuid.New().String()
	err := s.Link("hook", "", map[string]interface{}{"pin": pin})
	if err != nil {
		return "", err
	}
	username := base64.StdEncoding.EncodeToString([]byte(s.Client.Keychain.Username))
	encodedUsername := strings.TrimRight(username, "=")
	return fmt.Sprintf("%s/hook/%s/%s", s.Client.API, encodedUsername, pin), nil
}

func (s *AccountService) Add(item model.Account) error {
	return s.Link(s.Client.Username, item.Member, nil)
}

func (s *AccountService) Delete(item model.Account) error {
	return s.Unlink(s.Client.Username, item.Member)
}

func (s *AccountService) Unlink(username, id string) error {
	return s.linkHelper("DELETE", username, id, nil)
}

func (s *AccountService) Link(username, id string, config map[string]interface{}) error {
	return s.linkHelper("POST", username, id, config)
}
