package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

type Chariot struct {
	*Keychain

	Assets     *AssetService
	Attributes *AttributeService
	Risks      *RiskService
	Files      *FileService
	Jobs       *JobService
	Accounts   *AccountService
}

func NewClient(profile string) *Chariot {
	chariot := &Chariot{
		Keychain: NewKeychainFromIniFile(profile),
	}

	chariot.Accounts = NewAccountService(chariot)
	chariot.Assets = NewAssetService(chariot)
	chariot.Attributes = NewAttributeService(chariot)
	chariot.Files = NewFileService(chariot)
	chariot.Jobs = NewJobService(chariot)
	chariot.Risks = NewRiskService(chariot)

	return chariot
}

func (c *Chariot) Search(term string) (*model.SearchResult, error) {
	baseURL, err := url.Parse(c.API + "/my")
	if err != nil {
		return nil, err
	}

	baseURL.RawQuery = url.Values{"key": {term}}.Encode()
	body, err := c.request("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	var search model.SearchResult
	err = json.Unmarshal(body, &search)
	if err != nil {
		return nil, err
	}

	return &search, nil
}

func (c *Chariot) request(method, url string, data []byte) ([]byte, error) {
	token, err := c.GetToken()
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	var req *http.Request
	if data != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(data))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	if c.GetAccount() != "" {
		req.Header.Add("account", c.GetAccount())
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d, %s", resp.StatusCode, body)
	}

	return body, nil
}
