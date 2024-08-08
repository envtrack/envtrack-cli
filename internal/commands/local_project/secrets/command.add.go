package secrets

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func addSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add new secret(s) to an environment",
		RunE:  runAddSecret,
	}
	cmd.Flags().StringP("environment", "e", "", "Environment name or shortname (required if no environment is selected)")
	cmd.Flags().StringToStringP("secret", "s", nil, "Secret(s) to add in key=value format (can be used multiple times)")
	cmd.Flags().Bool("local", false, "Add secret(s) locally without syncing to server")
	cmd.Flags().BoolP("overwrite", "w", false, "Overwrite secret if it already exists")
	return cmd
}

func runAddSecret(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	envName, _ := cmd.Flags().GetString("environment")
	secrets, _ := cmd.Flags().GetStringToString("secret")
	local, _ := cmd.Flags().GetBool("local")
	overwrite, _ := cmd.Flags().GetBool("overwrite")

	selectedEnv, err := getSelectedEnvironment(localCfg, envName)
	if err != nil {
		return err
	}

	if len(secrets) == 0 {
		return fmt.Errorf("no secrets provided. Use --secret flag (e.g., --secret KEY=VALUE)")
	}

	for secretName, secretValue := range secrets {
		err := addSingleSecret(selectedEnv, secretName, secretValue, local, localCfg, overwrite)
		if err != nil {
			return err
		}
	}

	err = config.LocalConf.SaveLocalConfig(*localCfg)
	if err != nil {
		return fmt.Errorf("error saving local configuration: %v", err)
	}

	fmt.Printf("Secret(s) added successfully to environment '%s'.\n", selectedEnv.Name)
	return nil
}

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

func addSingleSecret(env *config.LocalConfigEnvironment, name, value string, local bool, localCfg *config.LocalConfigParams, overwrite bool) error {
	// Generate unique key and hash
	// uniqueKey := fmt.Sprintf("%s:%s:%s:%s", localCfg.Organization.ID, localCfg.Project.ID, env.ID, name)
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

	if index != -1 {
		if !overwrite {
			return fmt.Errorf("secret '%s' found in this environment. Use --overwrite to overwrite secrets", name)
		}
	}

	newSecret := config.LocalConfigSecret{
		Name: name,
	}

	if !local {

		// Store secret value using SetSecret
		hash, err := config.LocalConf.SetSecretWithSimpleKey(name, value)
		if err != nil {
			return fmt.Errorf("error storing secret '%s': %v", name, err)
		}
		// Sync with server
		token, _ := config.GlobalConf.GetAuthToken()
		client := api.NewClient(token)
		_, err = client.CreateSecret(localCfg.Organization.ID, localCfg.Project.ID, env.ID, hash, value)
		if err != nil {
			return fmt.Errorf("error creating secret '%s' on server: %v", name, err)
		}

		newSecret.Value = hash

	} else {
		if index != -1 {
			err = config.LocalConf.DeleteSecret(name)
			if err != nil {
				return fmt.Errorf("error deleting secret '%s': %v", name, err)
			}
			env.Secrets = append(env.Secrets[:index], env.Secrets[index+1:]...)
		}
		uniqueKey, err = config.LocalConf.SetSecretWithSimpleKey(name, value)
		if err != nil {
			return fmt.Errorf("error storing secret '%s': %v", name, err)
		}
		newSecret.Value = uniqueKey
		env.Secrets = append(env.Secrets, newSecret)
	}

	return nil
}
