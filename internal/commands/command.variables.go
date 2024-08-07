package commands

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func variablesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "variables",
		Aliases: []string{"var"},
		GroupID: "data",
		Short:   "List variables for a given environment",
		Run:     requireAuth(runVariables),
	}
	cmd.Flags().StringP("organization", "o", "", "ID or shortname of the organization")
	cmd.Flags().StringP("project", "p", "", "ID or shortname of the project")
	cmd.Flags().StringP("environment", "e", "", "ID or shortname of the environment")
	cmd.MarkFlagRequired("organization")
	cmd.MarkFlagRequired("project")
	cmd.MarkFlagRequired("environment")
	return cmd
}

func runVariables(cmd *cobra.Command, args []string) {
	orgID, _ := cmd.Flags().GetString("organization")
	projID, _ := cmd.Flags().GetString("project")
	envID, _ := cmd.Flags().GetString("environment")

	token, _ := config.GlobalConf.GetAuthToken() // Error already checked in requireAuth
	client := api.NewClient(token)

	orgs, err := client.GetVariables(orgID, projID, envID)
	if err != nil {
		fmt.Printf("Error fetching organizations: %v\n", err)
		return
	}

	formatter, err := getFormatter(cmd.Context())
	if err != nil {
		fmt.Printf("Error getting formatter: %v\n", err)
		return
	}
	formattedOutput, _ := formatter.Format(orgs)
	fmt.Print(formattedOutput)
}
