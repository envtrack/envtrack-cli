package env

import (
	"fmt"
	"strings"

	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func listEnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List environments for the local project",
		RunE:  runListEnvCommand,
	}

	cmd.Flags().StringP("environment", "e", "", "Environment name or shortname (required if no environment is selected)\n")
	cmd.Flags().BoolP("reload", "r", false, "Reload variables from the server"+
		"Available fields: name, shortname, isSelected, variablesCount, secretsCount\n"+
		"String fields will use contains, boolean fields will use exact match, integer fields will use greater than or equal\n"+
		"Example: envtrack ctx env list --filter name=production")

	// Filters
	cmd.Flags().StringToString("filter", nil, "Filter environments (key=value pairs, can be used multiple times)")

	return cmd
}

func runListEnvCommand(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil || localCfg.Organization.ID == "" || localCfg.Project.ID == "" {
		fmt.Println("No local context found. Use 'envtrack init' to initialize a local project.")
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	filterValues, error := cmd.Flags().GetStringToString("filter")
	if error != nil {
		return error
	}
	filters := common.FlagFilter{Value: filterValues}
	filters.
		StringF("name").
		StringF("shortname").
		BoolF("isSelected").
		IntF("variablesCount").
		IntF("secretsCount")

	if err := filters.ValidateFilters(); err != nil {
		return err
	}

	nameFilter, _ := filters.GetString("name")
	shortnameFilter, _ := filters.GetString("shortname")
	selectedFilter, _ := filters.GetBool("isSelected")
	minVariables, _ := filters.GetInt("variablesCount")
	minSecrets, _ := filters.GetInt("secretsCount")

	if len(localCfg.Environments) > 0 {
		output := make(map[string][]map[string]interface{})
		output["environments"] = []map[string]interface{}{}
		for _, env := range localCfg.Environments {
			if (nameFilter == "" || strings.Contains(strings.ToLower(env.Name), strings.ToLower(nameFilter))) &&
				(shortnameFilter == "" || strings.Contains(strings.ToLower(env.ShortName), strings.ToLower(shortnameFilter))) &&
				(!selectedFilter || env.IsSelected) &&
				(len(env.Variables) >= minVariables) &&
				(len(env.Secrets) >= minSecrets) {
				output["environments"] = append(output["environments"], map[string]interface{}{
					"name":           env.Name,
					"shortName":      env.ShortName,
					"id":             env.ID,
					"isSelected":     env.IsSelected,
					"variablesCount": len(env.Variables),
					"secretsCount":   len(env.Secrets),
				})
			}
		}

		formatter, err := common.GetFormatter(cmd.Context())
		if err != nil {
			fmt.Printf("Error getting formatter: %v\n", err)
			return err
		}
		formattedOutput, _ := formatter.Format(output)
		fmt.Print(formattedOutput)
	} else {
		fmt.Println("No environments found. Use 'envtrack ctx env --reload' to load environments from the server.")
	}
	return nil
}
