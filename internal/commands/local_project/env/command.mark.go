package env

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func markEnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark <environment>",
		Short: "Mark an environment as selected",
		RunE:  runMarkEnvCommand,
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	}

	return cmd
}

func runMarkEnvCommand(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("environment name is required")
	}

	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil || localCfg.Organization.ID == "" || localCfg.Project.ID == "" {
		fmt.Println("No local context found. Use 'envtrack init' to initialize a local project.")
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	if len(localCfg.Environments) > 0 {
		for i, env := range localCfg.Environments {
			if env.ShortName == args[0] {
				localCfg.SelectedEnv = env.ShortName
				localCfg.Environments[i].IsSelected = true
				err := config.LocalConf.SaveLocalConfig(*localCfg)
				if err != nil {
					fmt.Printf("Error saving local configuration: %v\n", err)
					return err
				}
				output := fmt.Sprintf("Environment '%s' marked as selected.", args[0])
				formatter, err := common.GetFormatter(cmd.Context())
				if err != nil {
					fmt.Printf("Error getting formatter: %v\n", err)
					return err
				}
				formattedOutput, _ := formatter.Format(output)
				fmt.Print(formattedOutput)
				return nil
			}
		}
		return fmt.Errorf("environment '%s' not found", args[0])
	}
	return fmt.Errorf("no environments found. Use 'envtrack ctx env --reload' to load environments from the server")
}
