package commands

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func environmentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environments",
		Aliases: []string{"env"},
		GroupID: "data",
		Short:   "List environments for a given project",
		Run:     requireAuth(runEnvironments),
	}
	cmd.Flags().StringP("organization", "o", "", "ID or shortname of the organization")
	cmd.Flags().StringP("project", "p", "", "ID or shortname of the project")
	cmd.MarkFlagRequired("organization")
	cmd.MarkFlagRequired("project")
	return cmd
}

func runEnvironments(cmd *cobra.Command, args []string) {
	orgID, _ := cmd.Flags().GetString("organization")
	projID, _ := cmd.Flags().GetString("project")

	token, _ := config.GlobalConf.GetAuthToken() // Error already checked in requireAuth
	client := api.NewClient(token)

	orgs, err := client.GetEnvironments(orgID, projID)
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
