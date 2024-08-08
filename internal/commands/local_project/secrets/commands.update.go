package secrets

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func updateSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update or create secret(s) in an environment",
		RunE:  runUpdateSecret,
	}
	cmd.Flags().StringP("environment", "e", "", "Environment name or shortname (required if no environment is selected)")
	cmd.Flags().StringToStringP("secret", "s", nil, "Secret(s) to update or create in key=value format (can be used multiple times)")
	cmd.Flags().BoolP("overwrite", "w", false, "Create new secrets if they don't exist")
	cmd.Flags().Bool("local", false, "Update secret(s) locally without syncing to server")
	return cmd
}

func runUpdateSecret(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	envName, _ := cmd.Flags().GetString("environment")
	secrets, _ := cmd.Flags().GetStringToString("secret")
	overwrite, _ := cmd.Flags().GetBool("overwrite")
	local, _ := cmd.Flags().GetBool("local")

	selectedEnv, err := getSelectedEnvironment(localCfg, envName)
	if err != nil {
		return err
	}

	if len(secrets) == 0 {
		return fmt.Errorf("no secrets provided. Use --secret flag (e.g., --secret KEY=VALUE)")
	}

	for secretName, secretValue := range secrets {
		err := updateSingleSecret(selectedEnv, secretName, secretValue, overwrite, local, localCfg)
		if err != nil {
			return err
		}
	}

	err = config.LocalConf.SaveLocalConfig(*localCfg)
	if err != nil {
		return fmt.Errorf("error saving local configuration: %v", err)
	}

	fmt.Printf("Secret(s) updated successfully in environment '%s'.\n", selectedEnv.Name)
	return nil
}

func updateSingleSecret(env *config.LocalConfigEnvironment, name, value string, overwrite, local bool, localCfg *config.LocalConfigParams) error {
	index := -1
	uniqueKey := ""
	var err error
	for i, s := range env.Secrets {
		if s.Name == name {
			index = i
			uniqueKey = s.Value
			break
		}
	}

	if index == -1 {
		if !overwrite {
			return fmt.Errorf("secret '%s' not found in this environment. Use --overwrite to create new secrets", name)
		} else {
			// Update or create the secret in the keyring
			uniqueKey, err = config.LocalConf.SetSecretWithSimpleKey(name, value)
			if err != nil {
				return fmt.Errorf("error storing secret '%s' in local storage: %v", name, err)
			}

		}
	}

	newSecret := config.LocalConfigSecret{
		Name:  name,
		Value: uniqueKey,
	}

	if !local {
		// Sync with server
		token, _ := config.GlobalConf.GetAuthToken()
		client := api.NewClient(token)
		// var serverSecret *api.Secret
		var err error
		if index == -1 {
			_, err = client.CreateSecret(localCfg.Organization.ID, localCfg.Project.ID, env.ID, name, value)
		}
		// else {
		// _, err = client.UpdateSecret(localCfg.Organization.ID, localCfg.Project.ID, env.ID, name, value)
		// }
		if err != nil {
			return fmt.Errorf("error updating secret '%s' on server: %v", name, err)
		}
		// newSecret = config.LocalConfigSecret{
		// 	Name: serverSecret.Name,
		// 	Hash: hash,
		// }
	}

	if index == -1 {
		env.Secrets = append(env.Secrets, newSecret)
	} else {
		env.Secrets[index] = newSecret
	}

	return nil
}
