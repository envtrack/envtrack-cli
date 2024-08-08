package variables

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func listVariablesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List variables for an environment",
		RunE:  runListVariables,
	}

	cmd.Flags().StringP("environment", "e", "", "Environment name or shortname (required if no environment is selected)")
	cmd.Flags().BoolP("reload", "r", false, "Reload variables from the server")

	return cmd
}

func runListVariables(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	envName, _ := cmd.Flags().GetString("environment")
	reload, _ := cmd.Flags().GetBool("reload")

	var selectedEnv *config.LocalConfigEnvironment

	if localCfg.SelectedEnv != "" {
		for _, env := range localCfg.Environments {
			if env.Name == localCfg.SelectedEnv || env.ShortName == localCfg.SelectedEnv {
				selectedEnv = &env
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
				selectedEnv = &env
				break
			}
		}
		if selectedEnv == nil {
			return fmt.Errorf("environment '%s' not found", envName)
		}
	}

	if reload {
		// Reload variables from server
		token, _ := config.GlobalConf.GetAuthToken()
		client := api.NewClient(token)
		serverVars, err := client.GetVariables(localCfg.Organization.ID, localCfg.Project.ID, selectedEnv.ID)
		if err != nil {
			return fmt.Errorf("error fetching variables from server: %v", err)
		}
		selectedEnv.Variables = []config.LocalConfigVariable{}
		for _, v := range serverVars {
			selectedEnv.Variables = append(selectedEnv.Variables, config.LocalConfigVariable{
				Name:  v.Name,
				Value: v.Value,
			})
		}
		config.LocalConf.SaveLocalConfig(*localCfg)
	}

	formatter, err := common.GetFormatter(cmd.Context())
	if err != nil {
		return fmt.Errorf("error getting formatter: %v", err)
	}

	formattedOutput, _ := formatter.Format(selectedEnv.Variables)
	fmt.Print(formattedOutput)

	return nil
}
