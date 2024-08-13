package varfile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/envtrack/envtrack-cli/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func readVarfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read and display the content of a linked variable file",
		RunE:  runReadVarfile,
	}
	cmd.Flags().StringP("alias", "a", "", "Alias of the linked file to read (required)")
	cmd.Flags().BoolP("raw", "r", false, "Flatten the object structure")

	cmd.MarkFlagRequired("alias")
	return cmd
}

func runReadVarfile(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	env, err := localCfg.GetSelectedEnvironment()
	if err != nil {
		return fmt.Errorf("no environment selected. Please use --environment flag or select an environment")
	}

	alias, _ := cmd.Flags().GetString("alias")
	raw, _ := cmd.Flags().GetBool("raw")

	// Find the linked file by alias
	var filePath string
	for _, linkedFile := range env.LinkedFiles {
		if linkedFile.Alias == alias {
			filePath = linkedFile.Path
			break
		}
	}

	if filePath == "" {
		return fmt.Errorf("no linked file found with alias '%s'", alias)
	}

	// Read file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	formatter, err := common.GetFormatter(cmd.Context())
	if err != nil {
		return fmt.Errorf("error getting formatter: %v", err)
	}

	// Determine file type and parse
	fileType := determineFileType(content)
	switch fileType {
	case "json":
		var jsonData interface{}
		err = json.Unmarshal(content, &jsonData)
		if err != nil {
			return fmt.Errorf("error parsing JSON: %v", err)
		}

		outputData(formatter, jsonData, raw)
	case "yaml":
		var yamlData interface{}
		err = yaml.Unmarshal(content, &yamlData)
		if err != nil {
			return fmt.Errorf("error parsing YAML: %v", err)
		}

		outputData(formatter, yamlData, raw)
	default:
		fmt.Printf("File type: Unknown\nContent:\n%s\n", string(content))
	}

	return nil
}

func outputData(formatter output.Formatter, data interface{}, raw bool) {
	if raw {
		flattenedData := make(map[string]interface{})
		flatten(data, "", flattenedData)
		formattedOutput, _ := formatter.Format(flattenedData)
		fmt.Print(formattedOutput)
	} else {
		formattedOutput, _ := formatter.Format(data)
		fmt.Print(formattedOutput)
	}
}

func flatten(data interface{}, prefix string, result map[string]interface{}) {
	rt := reflect.TypeOf(data)
	rv := reflect.ValueOf(data)

	switch rt.Kind() {
	case reflect.Map:
		for _, key := range rv.MapKeys() {
			newPrefix := prefix
			if prefix != "" {
				newPrefix += "."
			}
			newPrefix += key.String()
			flatten(rv.MapIndex(key).Interface(), newPrefix, result)
		}
	case reflect.Slice, reflect.Array:
		result[prefix] = data
	default:
		result[prefix] = data
	}
}

func determineFileType(content []byte) string {
	// Try to parse as JSON
	var jsonTest json.RawMessage
	if json.Unmarshal(content, &jsonTest) == nil {
		return "json"
	}

	// Try to parse as YAML
	var yamlTest interface{}
	if yaml.Unmarshal(content, &yamlTest) == nil {
		return "yaml"
	}

	// If neither, return unknown
	return "unknown"
}

// Don't forget to add this new command to the varfile command group
func NewVarfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "varfile",
		Short: "Manage variable files",
	}

	cmd.AddCommand(linkVarfileCommand())
	cmd.AddCommand(readVarfileCommand()) // Add the new read command

	return cmd
}
