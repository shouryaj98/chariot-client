package sdk

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

func (s *Service[T]) List() ([]T, error) {
	var t T
	baseURL, err := url.Parse(s.Client.API + "/my")
	if err != nil {
		return nil, err
	}

	baseURL.RawQuery = url.Values{"key": {"#" + s.KeyName()}}.Encode()

	var allResults []T
	for {
		body, err := s.Client.request("GET", baseURL.String(), nil)
		if err != nil {
			return nil, err
		}

		var search model.SearchResult
		err = json.Unmarshal(body, &search)
		if err != nil {
			return nil, err
		}

		var result interface{}
		switch any(t).(type) {
		case model.Asset:
			result = search.Assets
		case model.Attribute:
			result = search.Attributes
		case model.Risk:
			result = search.Risks
		case model.File:
			result = search.Files
		case model.Job:
			result = search.Jobs
		case model.Account:
			result = search.Accounts
		default:
			return nil, fmt.Errorf("unsupported type %T", t)
		}

		resultSlice := reflect.ValueOf(result)
		genericSlice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(t)), resultSlice.Len(), resultSlice.Len())
		for i := 0; i < resultSlice.Len(); i++ {
			genericSlice.Index(i).Set(resultSlice.Index(i))
		}
		allResults = append(allResults, genericSlice.Interface().([]T)...)

		if offsetKey, ok := search.Offset["key"]; ok && offsetKey != "" {
			offsetString, _ := json.Marshal(search.Offset)
			baseURL.RawQuery = url.Values{
				"key":    {"#" + s.KeyName()},
				"offset": {string(offsetString)},
			}.Encode()
		} else {
			break
		}
	}

	return allResults, nil
}

func (s *Service[T]) Count() (int, error) {
	items, err := s.List()
	if err != nil {
		return 0, err
	}
	return len(items), nil
}
