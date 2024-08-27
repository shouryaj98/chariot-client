package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

func (c *Chariot) UploadPoE(risk model.Risk, poe []byte) error {
	return c.Upload(fmt.Sprintf("proofs/%s/%s", risk.DNS, risk.Name), poe)
}

func (c *Chariot) UploadDefinition(risk model.Risk, definition []byte) error {
	return c.Upload(fmt.Sprintf("definitions/%s", risk.Name), definition)
}

func (c *Chariot) Upload(name string, data []byte) error {
	baseURL, err := url.Parse(c.API + "/file")
	if err != nil {
		return err
	}

	baseURL.RawQuery = url.Values{"name": {name}}.Encode()

	resp, err := c.request("PUT", baseURL.String(), data)
	if err != nil {
		return err
	}

	var presigned struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(resp, &presigned); err != nil {
		return err
	}
	req, _ := http.NewRequest("PUT", presigned.URL, bytes.NewReader(data))
	client := &http.Client{}
	upResp, _ := client.Do(req)
	defer upResp.Body.Close()
	if upResp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("unexpected status code: %d", upResp.StatusCode)
}
