package variables

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func addVariableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add new variable(s) to an environment",
		RunE:  runAddVariable,
	}
	cmd.Flags().StringP("environment", "e", "", "Environment name or shortname (required if no environment is selected)")
	cmd.Flags().StringToStringP("var", "v", nil, "Variable(s) to add in key=value format (can be used multiple times)")
	cmd.Flags().Bool("local", false, "Add variable(s) locally without syncing to server")
	cmd.Flags().BoolP("overwrite", "w", false, "Overwrite variable if it already exists")
	return cmd
}

func runAddVariable(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	envName, _ := cmd.Flags().GetString("environment")
	variables, _ := cmd.Flags().GetStringToString("var")
	local, _ := cmd.Flags().GetBool("local")
	overwrite, _ := cmd.Flags().GetBool("overwrite")

	selectedEnv, err := getSelectedEnvironment(localCfg, envName)
	if err != nil {
		return err
	}

	if len(variables) == 0 {
		return fmt.Errorf("no variables provided. Use --var flag (e.g., --var KEY=VALUE)")
	}

	for varName, varValue := range variables {
		err := addSingleVariable(selectedEnv, varName, varValue, local, localCfg, overwrite)
		if err != nil {
			return err
		}
	}

	err = config.LocalConf.SaveLocalConfig(*localCfg)
	if err != nil {
		return fmt.Errorf("error saving local configuration: %v", err)
	}

	fmt.Printf("Variable(s) added successfully to environment '%s'.\n", selectedEnv.Name)
	return nil
}

// TODO: use common function
func getSelectedEnvironment(localCfg *config.LocalConfigParams, envName string) (*config.LocalConfigEnvironment, error) {
	if localCfg.SelectedEnv != "" {
		for i, env := range localCfg.Environments {
			if env.Name == localCfg.SelectedEnv || env.ShortName == localCfg.SelectedEnv {
				return &localCfg.Environments[i], nil
			}
		}
	}

	if envName == "" {
		return nil, fmt.Errorf("no environment selected. Please use --environment flag or select an environment")
	}

	for i, env := range localCfg.Environments {
		if env.Name == envName || env.ShortName == envName {
			return &localCfg.Environments[i], nil
		}
	}

	return nil, fmt.Errorf("environment '%s' not found", envName)
}

func addSingleVariable(env *config.LocalConfigEnvironment, name, value string, local bool, localCfg *config.LocalConfigParams, overwrite bool) error {
	// Check if variable already exists
	if !overwrite {
		for _, v := range env.Variables {
			if v.Name == name {
				return fmt.Errorf("variable '%s' already exists in this environment", name)
			}
		}
	} else {
		// remove existing variable
		for i, v := range env.Variables {
			if v.Name == name {
				env.Variables = append(env.Variables[:i], env.Variables[i+1:]...)
				break
			}
		}
	}

	newVar := config.LocalConfigVariable{
		Name:  name,
		Value: value,
	}

	if !local {
		// Sync with server
		token, _ := config.GlobalConf.GetAuthToken()
		client := api.NewClient(token)
		serverVar, err := client.CreateVariable(localCfg.Organization.ID, localCfg.Project.ID, env.ID, name, value)
		if err != nil {
			return fmt.Errorf("error creating variable '%s' on server: %v", name, err)
		}
		newVar = config.LocalConfigVariable{
			Name:  serverVar.Name,
			Value: serverVar.Value,
		}
	}

	env.Variables = append(env.Variables, newVar)
	return nil
}
