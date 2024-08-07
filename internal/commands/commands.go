package commands

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func requireAuth(run func(cmd *cobra.Command, args []string)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		token, err := config.GlobalConf.GetAuthToken()
		if err != nil || token == "" {
			fmt.Println("Authentication required. Please run 'envtrack auth <token>' to authenticate.")
			return
		}
		run(cmd, args)
	}
}

func AddCommands(rootCmd *cobra.Command) {
	rootCmd.AddGroup(
		&cobra.Group{ID: "auth", Title: "Authentication commands"},
		&cobra.Group{ID: "conf", Title: "Configuration commands"},
		&cobra.Group{ID: "data", Title: "Data commands"},
		&cobra.Group{ID: "local_project", Title: "Local project commands"},
		&cobra.Group{ID: "help", Title: "Help commands"},
	)
	rootCmd.AddCommand(
		authCommand(),
		organizationsCommand(),
		projectsCommand(),
		environmentsCommand(),
		variablesCommand(),
		configureCommand(),
		versionCommand(),
		initCommand(),
	)
}
