package secrets

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func getSecretsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name>",
		Args:  cobra.ExactArgs(1),
		Short: "Get secrets for the selected environment",
		RunE:  runGetSecrets,
	}
	cmd.Flags().StringP("environment", "e", "", "Environment name or shortname (required if no environment is selected)")
	cmd.Flags().BoolP("reload", "r", false, "Reload secrets from the server")
	return cmd
}

func runGetSecrets(cmd *cobra.Command, args []string) error {
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

	// Prepare a list of secret names (without values) for display
	secret := map[string]string{}
	for _, s := range selectedEnv.Secrets {
		if s.Name == args[0] {
			secretVal, err := config.LocalConf.GetSecret(s.Value)
			if err != nil {
				return fmt.Errorf("error getting secret value: %v", err)
			}
			secret[s.Name] = secretVal
			break
		}
	}

	formatter, err := common.GetFormatter(cmd.Context())
	if err != nil {
		return fmt.Errorf("error getting formatter: %v", err)
	}

	formattedOutput, _ := formatter.Format(secret)
	fmt.Print(formattedOutput)

	return nil
}
