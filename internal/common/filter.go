package common

import (
	"fmt"
	"strconv"
)

type flagFilterType string

const (
	stringFilter flagFilterType = "string"
	boolFilter   flagFilterType = "bool"
	intFilter    flagFilterType = "int"
)

type FlagFilter struct {
	Value map[string]string

	StringKeys []string
	BoolKeys   []string
	IntKeys    []string

	validKeysMap map[string]flagFilterType
}

func (f *FlagFilter) StringF(key string) *FlagFilter {
	f.StringKeys = append(f.StringKeys, key)
	return f
}

func (f *FlagFilter) BoolF(key string) *FlagFilter {
	f.BoolKeys = append(f.BoolKeys, key)
	return f
}

func (f *FlagFilter) IntF(key string) *FlagFilter {
	f.IntKeys = append(f.IntKeys, key)
	return f
}

func (f *FlagFilter) parseValidKeys() {
	f.validKeysMap = make(map[string]flagFilterType)
	for _, key := range f.StringKeys {
		f.validKeysMap[key] = stringFilter
	}
	for _, key := range f.BoolKeys {
		f.validKeysMap[key] = boolFilter
	}
	for _, key := range f.IntKeys {
		f.validKeysMap[key] = intFilter
	}
}

func (f *FlagFilter) ValidateFilters() error {
	f.parseValidKeys()

	for key, value := range f.Value {
		if _, ok := f.validKeysMap[key]; !ok {
			return fmt.Errorf("invalid filter key: %s", key)
		}
		switch f.validKeysMap[key] {
		case stringFilter:
			// Do nothing
		case boolFilter:
			if _, err := strconv.ParseBool(value); err != nil {
				return fmt.Errorf("invalid boolean value for filter %s: %s", key, value)
			}
		case intFilter:
			if _, err := strconv.Atoi(value); err != nil {
				return fmt.Errorf("invalid integer value for filter %s: %s", key, value)
			}
		}
	}
	return nil
}

func (f *FlagFilter) GetString(key string) (string, bool) {
	if f.validKeysMap[key] != stringFilter {
		return "", false
	}
	return f.Value[key], true
}

func (f *FlagFilter) GetBool(key string) (bool, bool) {
	if f.validKeysMap[key] != boolFilter {
		return false, false
	}
	b, _ := strconv.ParseBool(f.Value[key])
	return b, true
}

func (f *FlagFilter) GetInt(key string) (int, bool) {
	if f.validKeysMap[key] != intFilter {
		return 0, false
	}
	i, _ := strconv.Atoi(f.Value[key])
	return i, true
}
