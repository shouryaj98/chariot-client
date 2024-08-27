package sdk

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

func (s *Service[T]) upsert(method string, item T) error {
	baseURL, err := url.Parse(s.Client.API + "/" + strings.TrimPrefix(s.KeyName(), "#"))
	if err != nil {
		return err
	}

	body, err := json.Marshal(item)
	if err != nil {
		return err
	}

	_, err = s.Client.request(method, baseURL.String(), body)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service[T]) Update(item T) error {
	return s.upsert("PUT", item)
}

func (s *Service[T]) Add(item T) error {
	return s.upsert("POST", item)
}

func (s *RiskService) Add(item model.Risk) error {
	asset := model.NewAsset(item.DNS, item.DNS)
	item.Key = asset.Key
	return s.Service.Add(item)
}

func (s *AttributeService) Add(item model.Attribute) error {
	parts := strings.Split(item.Key, "#")
	item.Key = "#" + strings.Join(parts[4:], "#")
	return s.Service.Add(item)
}

func (s *FileService) Delete(item model.File) error {
	baseURL, err := url.Parse(s.Client.API + "/" + s.KeyName())
	if err != nil {
		return err
	}
	baseURL.RawQuery = url.Values{"name": {item.Name}}.Encode()
	_, err = s.Client.request("DELETE", baseURL.String(), nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *AssetService) Delete(item model.Asset) error {
	item.Status = model.Deleted
	return s.Service.Update(item)
}

func (s *Service[T]) Delete(item T) error {
	baseURL, err := url.Parse(s.Client.API + "/" + s.KeyName())
	if err != nil {
		return err
	}

	var key string
	v := reflect.ValueOf(item)
	field := v.FieldByName("Key")
	if field.IsValid() && field.Kind() == reflect.String {
		key = field.String()
	} else {
		return fmt.Errorf("unsupported type %T", item)
	}

	baseURL.RawQuery = url.Values{"key": {key}}.Encode()
	deleteBody, _ := json.Marshal(map[string]string{"key": key})
	_, err = s.Client.request("DELETE", baseURL.String(), deleteBody)
	if err != nil {
		return err
	}
	return nil
}
