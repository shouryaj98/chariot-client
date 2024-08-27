package capabilities

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Crowdstrike struct {
	Job   model.Job
	Asset model.Asset
	XYZ
}

type FalconClient struct {
	Token      string
	BaseURL    string
	HttpClient *http.Client
}

type Device struct {
	HostID   string `json:"device_id"`
	Hostname string `json:"hostname"`
	OS       string `json:"platform_name"`
}

type Exclusion struct {
	ID     string `json:"id"`
	Path   string `json:"value"`
	Global bool   `json:"applied_globally"`
}

type GenericResponse[T any] struct {
	Resources []T `json:"resources"`
}

type Policy struct {
	ID       string               `json:"id"`
	Name     string               `json:"name"`
	Platform string               `json:"platform_name"`
	Settings []PreventionCategory `json:"prevention_settings"`
}

type PreventionCategory struct {
	Name     string              `json:"name"`
	Controls []PreventionControl `json:"settings"`
}

type PreventionControl struct {
	Name    string                 `json:"name"`
	Options map[string]interface{} `json:"value"`
}

func NewCrowdstrike(job model.Job) model.Capability {
	return &Crowdstrike{Asset: job.Target, Job: job, XYZ: NewXYZ()}
}

func (task *Crowdstrike) Match() bool {
	_, client := task.Asset.Config["clientID"]
	_, secret := task.Asset.Config["secret"]
	return client && secret
}

func (task *Crowdstrike) Invoke() error {
	c := task.Asset.Config

	client, err := NewFalconClient(c["clientID"], c["secret"], task.Asset.Name)
	if err != nil {
		return fmt.Errorf("failed to create connection: %v", err)
	}

	devices, err := client.Devices()
	if err != nil {
		return fmt.Errorf("failed to identify devices: %v", err)
	}

	for _, computer := range devices {
		task.Job.Stream <- model.NewAsset(computer.Hostname, fmt.Sprintf("%s", computer.HostID))
	}

	exclusions, err := client.Exclusions()
	if err != nil {
		return fmt.Errorf("failed to identify exclusions: %v", err)
	}

	for _, exclusion := range exclusions {
		if exclusion.isDangerous() {
			task.Asset.DNS = exclusion.ID
			task.Job.Stream <- model.NewRisk(task.Asset, "dangerous-exclusion-rule")
		}
	}

	policies, err := client.Policies()
	if err != nil {
		return fmt.Errorf("failed to identify policies: %v", err)
	}

	rules := []struct {
		Platform string
		Setting  string
		RiskName string
	}{
		{"Windows", "Ransomware", "windows-ransomware-control-disabled"},
		{"Linux", "Enhanced Visibility", "linux-prevention-control-disabled"},
	}

	for _, policy := range policies {
		if policy.Name == "Phase 1 - initial deployment" {
			continue
		}

		for _, settings := range policy.Settings {
			for _, rule := range rules {
				if policy.Platform == rule.Platform && settings.Name == rule.Setting {
					for _, control := range settings.Controls {
						if !control.isEnabled() {
							task.Asset.DNS = hex.EncodeToString(sha256.New().Sum([]byte(fmt.Sprintf("%s:%s", policy.ID, control.Name))))
							task.Job.Stream <- model.NewRisk(task.Asset, rule.RiskName)
						}
					}
				}
			}
		}
	}

	return nil
}

func NewFalconClient(clientID, secret, baseURL string) (*FalconClient, error) {
	token, err := getToken(clientID, secret, baseURL)
	if err != nil {
		return nil, err
	}

	return &FalconClient{
		Token:      token,
		BaseURL:    baseURL,
		HttpClient: &http.Client{},
	}, nil
}

func fetchEntities[T any](c *FalconClient, name string) ([]T, error) {
	query := fmt.Sprintf("%s/policy/queries/%s/v1?limit=500", c.BaseURL, name)
	resp, err := sendRequest[GenericResponse[string]](c, "GET", query, []byte{})
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	for _, id := range resp.Resources {
		params.Add("ids", id)
	}

	get := fmt.Sprintf("%s/policy/entities/%s/v1?%s", c.BaseURL, name, params.Encode())
	entities, err := sendRequest[GenericResponse[T]](c, "GET", get, []byte{})
	if err != nil {
		return nil, err
	}

	return entities.Resources, nil
}

func (c *FalconClient) Devices() ([]Device, error) {
	url := fmt.Sprintf("%s/devices/queries/devices/v1", c.BaseURL)
	resp, err := sendRequest[GenericResponse[string]](c, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch hosts: %v", err)
	}

	ids := resp.Resources
	if len(ids) == 0 {
		return nil, nil
	}

	url = fmt.Sprintf("%s/devices/entities/devices/v1", c.BaseURL)
	body, err := json.Marshal(map[string]interface{}{"ids": ids})
	if err != nil {
		return nil, err
	}

	devices, err := sendRequest[GenericResponse[Device]](c, "GET", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to identify devices: %v", err)
	}

	return devices.Resources, nil
}

func (c *FalconClient) Exclusions() ([]Exclusion, error) {
	return fetchEntities[Exclusion](c, "sv-exclusions")
}

func (c *FalconClient) Policies() ([]Policy, error) {
	return fetchEntities[Policy](c, "prevention")
}

func sendRequest[T any](c *FalconClient, method, url string, body []byte) (T, error) {
	var result T

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(respBody, &result)
	return result, err
}

func getToken(clientID, secret, baseURL string) (string, error) {
	url := fmt.Sprintf("%s/oauth2/token", baseURL)
	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + secret))
	data := "grant_type=client_credentials"

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response.AccessToken, nil
}

func (e *Exclusion) isDangerous() bool {
	writablePaths := []string{"/tmp/", "/var/tmp/", "/private/tmp/", "C:\\Windows\\Temp\\", "C:\\Temp\\", "C:\\Users\\Public\\", "C:\\ProgramData\\"}

	for _, writablePath := range writablePaths {
		if strings.HasPrefix(strings.ToLower(e.Path), strings.ToLower(writablePath)) {
			return true
		}
	}

	return false
}

func (c *PreventionControl) isEnabled() bool {
	return c.Options["enabled"].(bool)
}
