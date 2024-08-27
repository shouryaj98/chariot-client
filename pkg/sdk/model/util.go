// provides helpers to build models based on keys or names
package model

import (
	"fmt"
	"strings"
)

func GetRiskFromKey(key string) (*Risk, error) {
	parts := strings.Split(key, "#")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid key format, expected #risk#dns#name, got %s", key)
	}
	risk := NewRisk(Asset{DNS: parts[2]}, parts[3])
	return &risk, nil
}

func GetAssetFromKey(key string) (*Asset, error) {
	parts := strings.Split(key, "#")
	var asset Asset
	switch len(parts) {
	case 3:
		asset = NewAsset(parts[2], parts[2])
	case 4:
		asset = NewAsset(parts[2], parts[3])
	default:
		return nil, fmt.Errorf("invalid key format, expected #asset#dns#name, or #asset#dns got %s", key)
	}
	return &asset, nil
}

func GetAttributeFromKey(key string) (*Attribute, error) {
	parts := strings.Split(key, "#")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid key format, expected #attribute#name#value#KEY, got %s", key)
	}
	attribute := NewAttribute(parts[2], parts[3], "#"+strings.Join(parts[4:], "#"))
	return &attribute, nil
}

func FilterAssetsByKey(assets []Asset, filter string) []Asset {
	filter = strings.ToLower(filter)
	filtered := make([]Asset, 0)
	for _, asset := range assets {
		key := strings.ToLower(asset.Key)
		if strings.Contains(key, filter) {
			filtered = append(filtered, asset)
		}
	}
	return filtered
}

func FilterFilesByName(files []File, filter string) []File {
	filter = strings.ToLower(filter)
	filtered := make([]File, 0)
	for _, file := range files {
		name := strings.ToLower(file.Name)
		if strings.Contains(name, filter) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}
