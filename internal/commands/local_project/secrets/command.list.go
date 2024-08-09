package secrets

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func listSecretsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List secrets for an environment",
		RunE:  runListSecrets,
	}
	cmd.Flags().StringP("environment", "e", "", "Environment name or shortname (required if no environment is selected)")
	cmd.Flags().BoolP("reload", "r", false, "Reload secrets from the server")
	return cmd
}

func runListSecrets(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	envName, _ := cmd.Flags().GetString("environment")
	// reload, _ := cmd.Flags().GetBool("reload")

	var selectedEnv *config.LocalConfigEnvironment
	if localCfg.SelectedEnv != "" {
		for _, env := range localCfg.Environments {
			if env.Name == localCfg.SelectedEnv || env.ShortName == localCfg.SelectedEnv {
				selectedEnv = env
				break
			}
		}
	}

	if selectedEnv == nil {
		if envName == "" {
			return fmt.Errorf("no environment selected. Please use --environment flag or select an environment")
		}
		for _, env := range localCfg.Environments {
			if env.Name == envName || env.ShortName == envName {
				selectedEnv = env
				break
			}
		}
		if selectedEnv == nil {
			return fmt.Errorf("environment '%s' not found", envName)
		}
	}

	// if reload {
	// 	// Reload secrets from server
	// 	token, _ := config.GlobalConf.GetAuthToken()
	// 	client := api.NewClient(token)
	// 	serverSecrets, err := client.GetSecrets(localCfg.Organization.ID, localCfg.Project.ID, selectedEnv.ID)
	// 	if err != nil {
	// 		return fmt.Errorf("error fetching secrets from server: %v", err)
	// 	}

	// 	selectedEnv.Secrets = []config.LocalConfigSecret{}
	// 	for _, s := range serverSecrets {
	// 		// Store secret value using keyring
	// 		err := config.LocalConf.SetSecret(localCfg.Organization.ID, localCfg.Project.ID, selectedEnv.ID, s.Name, s.Value)
	// 		if err != nil {
	// 			return fmt.Errorf("error storing secret: %v", err)
	// 		}
	// 		selectedEnv.Secrets = append(selectedEnv.Secrets, config.LocalConfigSecret{
	// 			Name: s.Name,
	// 		})
	// 	}
	// 	config.LocalConf.SaveLocalConfig(*localCfg)
	// }

	// Prepare a list of secret names (without values) for display
	secretNames := []map[string]string{}
	for _, s := range selectedEnv.Secrets {
		secretNames = append(secretNames, map[string]string{
			"name":  s.Name,
			"value": s.Value,
		})
	}

	formatter, err := common.GetFormatter(cmd.Context())
	if err != nil {
		return fmt.Errorf("error getting formatter: %v", err)
	}

	formattedOutput, _ := formatter.Format(secretNames)
	fmt.Print(formattedOutput)

	return nil
}
