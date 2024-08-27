package sdk

import (
	"fmt"
	"net/url"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

func (c *Chariot) DownloadPoE(risk model.Risk) ([]byte, error) {
	return c.Download(fmt.Sprintf("proofs/%s/%s", risk.DNS, risk.Name))
}

func (c *Chariot) DownloadDefinition(risk model.Risk) ([]byte, error) {
	return c.Download(fmt.Sprintf("definitions/%s", risk.Name))
}

func (c *Chariot) DownloadFile(file model.File) ([]byte, error) {
	return c.Download(file.Name)
}

func (c *Chariot) Download(name string) ([]byte, error) {
	baseURL, err := url.Parse(c.API + "/file")
	if err != nil {
		return nil, err
	}

	baseURL.RawQuery = url.Values{"name": {name}}.Encode()
	body, err := c.request("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return body, nil
}
