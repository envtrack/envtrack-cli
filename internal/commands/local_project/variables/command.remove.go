package variables

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func removeVariableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove variable(s) from an environment",
		RunE:  runRemoveVariable,
	}
	cmd.Flags().StringP("environment", "e", "", "Environment name or shortname (required if no environment is selected)")
	cmd.Flags().StringSliceP("name", "n", nil, "Name(s) of the variable(s) to remove")
	cmd.Flags().Bool("local", false, "Remove variable(s) locally without syncing to server")
	return cmd
}

func runRemoveVariable(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	envName, _ := cmd.Flags().GetString("environment")
	varNames, _ := cmd.Flags().GetStringSlice("name")
	local, _ := cmd.Flags().GetBool("local")

	selectedEnv, err := getSelectedEnvironment(localCfg, envName)
	if err != nil {
		return err
	}

	if len(varNames) == 0 {
		return fmt.Errorf("no variable names provided. Use --name flag (can be used multiple times)")
	}

	for _, varName := range varNames {
		err := removeSingleVariable(selectedEnv, varName, local, localCfg)
		if err != nil {
			return err
		}
	}

	err = config.LocalConf.SaveLocalConfig(*localCfg)
	if err != nil {
		return fmt.Errorf("error saving local configuration: %v", err)
	}

	fmt.Printf("Variable(s) removed successfully from environment '%s'.\n", selectedEnv.Name)
	return nil
}

func removeSingleVariable(env *config.LocalConfigEnvironment, name string, local bool, localCfg *config.LocalConfigParams) error {
	index := -1
	for i, v := range env.Variables {
		if v.Name == name {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("variable '%s' not found in this environment", name)
	}

	if !local {
		// Sync with server
		token, _ := config.GlobalConf.GetAuthToken()
		client := api.NewClient(token)
		err := client.DeleteVariable(localCfg.Organization.ID, localCfg.Project.ID, env.ID, name)
		if err != nil {
			return fmt.Errorf("error deleting variable '%s' on server: %v", name, err)
		}
	}

	// Remove the variable from the local config
	env.Variables = append(env.Variables[:index], env.Variables[index+1:]...)
	return nil
}
