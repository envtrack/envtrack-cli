package secrets

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func removeSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove secret(s) from an environment",
		RunE:  runRemoveSecret,
	}
	cmd.Flags().StringP("environment", "e", "", "Environment name or shortname (required if no environment is selected)")
	cmd.Flags().StringSliceP("name", "n", nil, "Name(s) of the secret(s) to remove")
	cmd.Flags().Bool("local", false, "Remove secret(s) locally without syncing to server")
	return cmd
}

func runRemoveSecret(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	envName, _ := cmd.Flags().GetString("environment")
	secretNames, _ := cmd.Flags().GetStringSlice("name")
	local, _ := cmd.Flags().GetBool("local")

	selectedEnv, err := getSelectedEnvironment(localCfg, envName)
	if err != nil {
		return err
	}

	if len(secretNames) == 0 {
		return fmt.Errorf("no secret names provided. Use --name flag (can be used multiple times)")
	}

	for _, secretName := range secretNames {
		err := removeSingleSecret(selectedEnv, secretName, local, localCfg)
		if err != nil {
			return err
		}
	}

	err = config.LocalConf.SaveLocalConfig(*localCfg)
	if err != nil {
		return fmt.Errorf("error saving local configuration: %v", err)
	}

	fmt.Printf("Secret(s) removed successfully from environment '%s'.\n", selectedEnv.Name)
	return nil
}

func removeSingleSecret(env *config.LocalConfigEnvironment, name string, local bool, localCfg *config.LocalConfigParams) error {
	index := -1
	currentSecret := &config.LocalConfigSecret{}
	for i, s := range env.Secrets {
		if s.Name == name {
			index = i
			currentSecret = s
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("secret '%s' not found in this environment", name)
	}

	if !local {
		// Sync with server
		token, _ := config.GlobalConf.GetAuthToken()
		client := api.NewClient(token)
		err := client.DeleteSecret(localCfg.Organization.ID, localCfg.Project.ID, env.ID, name)
		if err != nil {
			return fmt.Errorf("error deleting secret '%s' on server: %v", name, err)
		}
	}

	// Remove the secret from the local config
	env.Secrets = append(env.Secrets[:index], env.Secrets[index+1:]...)

	// Delete the secret value from the keyring
	err := config.LocalConf.DeleteSecret(currentSecret.Value)
	if err != nil {
		return fmt.Errorf("error deleting secret '%s' from local storage: %v", name, err)
	}

	return nil
}
