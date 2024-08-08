package commands

import (
	"github.com/envtrack/envtrack-cli/internal/commands/auth"
	"github.com/envtrack/envtrack-cli/internal/commands/conf"
	"github.com/envtrack/envtrack-cli/internal/commands/data"
	"github.com/envtrack/envtrack-cli/internal/commands/help"
	"github.com/envtrack/envtrack-cli/internal/commands/local_project"
	"github.com/spf13/cobra"
)

func AddCommands(rootCmd *cobra.Command) {
	rootCmd.AddGroup(
		&cobra.Group{ID: "auth", Title: "Authentication commands"},
		&cobra.Group{ID: "conf", Title: "Configuration commands"},
		&cobra.Group{ID: "data", Title: "Data commands"},
		&cobra.Group{ID: "local_project", Title: "Local project commands"},
		&cobra.Group{ID: "help", Title: "Help commands"},
	)
	rootCmd.AddCommand(
		auth.AuthCommand(),
		data.OrganizationsCommand(),
		data.ProjectsCommand(),
		data.EnvironmentsCommand(),
		data.VariablesCommand(),
		conf.ConfigureCommand(),
		help.VersionCommand(),
		local_project.LocalContextCommand(),
	)
}
