package commands

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func requireAuth(run func(cmd *cobra.Command, args []string)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		token, err := config.GetAuthToken()
		if err != nil || token == "" {
			fmt.Println("Authentication required. Please run 'envtrack auth <token>' to authenticate.")
			return
		}
		run(cmd, args)
	}
}

func AddCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(
		authCommand(),
		organizationsCommand(),
		projectsCommand(),
		environmentsCommand(),
		variablesCommand(),
	)
}

func organizationsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "organizations",
		Short: "List organizations",
		Run:   requireAuth(runOrganizations),
	}
}

func projectsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "List projects for a given organization",
		Run:   requireAuth(runProjects),
	}
	cmd.Flags().String("organizationId", "", "ID of the organization")
	cmd.MarkFlagRequired("organizationId")
	return cmd
}

func environmentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "environments",
		Short: "List environments for a given project",
		Run:   requireAuth(runEnvironments),
	}
	cmd.Flags().String("organizationId", "", "ID of the organization")
	cmd.Flags().String("projectId", "", "ID of the project")
	cmd.MarkFlagRequired("organizationId")
	cmd.MarkFlagRequired("projectId")
	return cmd
}

func variablesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variables",
		Short: "List variables for a given environment",
		Run:   requireAuth(runVariables),
	}
	cmd.Flags().String("organizationId", "", "ID of the organization")
	cmd.Flags().String("projectId", "", "ID of the project")
	cmd.Flags().String("environmentId", "", "ID of the environment")
	cmd.MarkFlagRequired("organizationId")
	cmd.MarkFlagRequired("projectId")
	cmd.MarkFlagRequired("environmentId")
	return cmd
}
