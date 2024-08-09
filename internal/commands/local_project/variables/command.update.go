package variables

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func updateVariableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update or create variable(s) in an environment",
		RunE:  runUpdateVariable,
	}
	cmd.Flags().StringP("environment", "e", "", "Environment name or shortname (required if no environment is selected)")
	cmd.Flags().StringToStringP("var", "v", nil, "Variable(s) to update or create in key=value format (can be used multiple times)")
	cmd.Flags().BoolP("overwrite", "w", false, "Create new variables if they don't exist")
	cmd.Flags().Bool("local", false, "Update variable(s) locally without syncing to server")
	return cmd
}

func runUpdateVariable(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	envName, _ := cmd.Flags().GetString("environment")
	variables, _ := cmd.Flags().GetStringToString("var")
	overwrite, _ := cmd.Flags().GetBool("overwrite")
	local, _ := cmd.Flags().GetBool("local")

	selectedEnv, err := getSelectedEnvironment(localCfg, envName)
	if err != nil {
		return err
	}

	if len(variables) == 0 {
		return fmt.Errorf("no variables provided. Use --var flag (e.g., --var KEY=VALUE)")
	}

	for varName, varValue := range variables {
		err := updateSingleVariable(selectedEnv, varName, varValue, overwrite, local, localCfg)
		if err != nil {
			return err
		}
	}

	err = config.LocalConf.SaveLocalConfig(*localCfg)
	if err != nil {
		return fmt.Errorf("error saving local configuration: %v", err)
	}

	fmt.Printf("Variable(s) updated successfully in environment '%s'.\n", selectedEnv.Name)
	return nil
}

func updateSingleVariable(env *config.LocalConfigEnvironment, name, value string, overwrite, local bool, localCfg *config.LocalConfigParams) error {
	index := -1
	for i, v := range env.Variables {
		if v.Name == name {
			index = i
			break
		}
	}

	if index == -1 && !overwrite {
		return fmt.Errorf("variable '%s' not found in this environment. Use --overwrite to create new variables", name)
	}

	newVar := &config.LocalConfigVariable{
		Name:  name,
		Value: value,
	}

	if !local {
		// Sync with server
		token, _ := config.GlobalConf.GetAuthToken()
		client := api.NewClient(token)
		var serverVar *api.Variable
		var err error
		if index == -1 {
			serverVar, err = client.CreateVariable(localCfg.Organization.ID, localCfg.Project.ID, env.ID, name, value)
		} else {
			serverVar, err = client.UpdateVariable(localCfg.Organization.ID, localCfg.Project.ID, env.ID, name, value)
		}
		if err != nil {
			return fmt.Errorf("error updating variable '%s' on server: %v", name, err)
		}
		newVar = &config.LocalConfigVariable{
			Name:  serverVar.Name,
			Value: serverVar.Value,
		}
	}

	if index == -1 {
		env.Variables = append(env.Variables, newVar)
	} else {
		env.Variables[index] = newVar
	}
	return nil
}
