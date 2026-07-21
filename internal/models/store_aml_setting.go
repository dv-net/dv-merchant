package models

import "github.com/goccy/go-json"

func (s StoreAmlSetting) ParseIgnoredSignalCategories() ([]string, error) {
	if len(s.IgnoredSignalCategories) == 0 {
		return nil, nil
	}
	var categories []string
	if err := json.Unmarshal(s.IgnoredSignalCategories, &categories); err != nil {
		return nil, err
	}
	return categories, nil
}
