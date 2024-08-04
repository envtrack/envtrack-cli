package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

type Formatter interface {
	Format(data interface{}) (string, error)
}

type JSONFormatter struct{}

func (f JSONFormatter) Format(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

type YAMLFormatter struct{}

func (f YAMLFormatter) Format(data interface{}) (string, error) {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

type CSVFormatter struct{}

// Update CSVFormatter to handle various data types
func (f CSVFormatter) Format(data interface{}) (string, error) {
	var records [][]string

	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Slice {
		records = make([][]string, value.Len()+1)
		if value.Len() > 0 {
			// Get headers from the first item's field names
			firstItem := value.Index(0)
			headers := make([]string, firstItem.NumField())
			for i := 0; i < firstItem.NumField(); i++ {
				headers[i] = firstItem.Type().Field(i).Name
			}
			records[0] = headers

			// Get values for each item
			for i := 0; i < value.Len(); i++ {
				item := value.Index(i)
				record := make([]string, item.NumField())
				for j := 0; j < item.NumField(); j++ {
					record[j] = fmt.Sprintf("%v", item.Field(j).Interface())
				}
				records[i+1] = record
			}
		}
	} else {
		return "", fmt.Errorf("data must be a slice for CSV formatting")
	}

	var buf strings.Builder
	writer := csv.NewWriter(&buf)
	err := writer.WriteAll(records)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

type BashFormatter struct{}

// Implement the Format method for BashFormatter
// should return lines of export VARNAME=VALUE
// works with key value structs, maps, and slices of key value structs
func (f BashFormatter) Format(data interface{}) (string, error) {
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Slice {
		var buf strings.Builder
		for i := 0; i < value.Len(); i++ {
			item := value.Index(i)
			if item.Kind() == reflect.Struct {
				if item.Type().Field(0).Name == "Name" && item.Type().Field(1).Name == "Value" {
					fmt.Fprintf(&buf, "export %s=%q\n", item.Field(0).Interface(), item.Field(1).Interface())
				} else if item.Type().Field(0).Name == "Value" && item.Type().Field(1).Name == "Name" {
					fmt.Fprintf(&buf, "export %s=%q\n", item.Field(1).Interface(), item.Field(0).Interface())
				} else {
					for j := 0; j < item.NumField(); j++ {
						fmt.Fprintf(&buf, "export %s=%q\n", item.Type().Field(j).Name, item.Field(j).Interface())
					}
				}
			} else {
				return "", fmt.Errorf("data must be a slice of structs for Bash formatting")
			}
		}
		return buf.String(), nil
	}

	if value.Kind() == reflect.Map {
		var buf strings.Builder
		for _, key := range value.MapKeys() {
			fmt.Fprintf(&buf, "export %s=%q\n", key.Interface(), value.MapIndex(key).Interface())
		}
		return buf.String(), nil
	}

	return "", fmt.Errorf("data must be a slice or map for Bash formatting")
}

func GetFormatter(format string) (Formatter, error) {
	switch format {
	case "json":
		return JSONFormatter{}, nil
	case "yaml":
		return YAMLFormatter{}, nil
	case "csv":
		return CSVFormatter{}, nil
	case "bash":
		return BashFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
