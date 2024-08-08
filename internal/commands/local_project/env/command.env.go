package env

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func LocalEnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environments <name>",
		Aliases: []string{"env"},
		Short:   "Display local environment information",
		RunE:    runLocalEnvCommand,
	}
	cmd.Flags().BoolP("reload", "r", false, "Reload local environments from the server.")

	cmd.AddCommand(addEnvCommand())
	cmd.AddCommand(listEnvCommand())
	cmd.AddCommand(markEnvCommand())
	return cmd
}

func runLocalEnvCommand(cmd *cobra.Command, args []string) error {
	if len(args) > 0 && args[0] != "" {
		localCfg, err := config.LocalConf.GetLocalConfig()
		if err != nil || localCfg.Organization.ID == "" || localCfg.Project.ID == "" {
			fmt.Println("No local context found. Use 'envtrack init' to initialize a local project.")
			return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
		}
		for _, env := range localCfg.Environments {
			if env.ShortName == args[0] {
				output := map[string]interface{}{
					"name":           env.Name,
					"shortName":      env.ShortName,
					"id":             env.ID,
					"isSelected":     env.IsSelected,
					"variablesCount": len(env.Variables),
					"secretsCount":   len(env.Secrets),
				}
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
		fmt.Printf("Environment '%s' not found.\n", args[0])
		return fmt.Errorf("environment '%s' not found", args[0])
	}
	return fmt.Errorf("environment name is required")
}
